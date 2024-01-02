package bestpractices

import (
	"context"
	"encoding/json"
	"io/fs"
	"strings"

	"github.com/mansam/policy-runner/k8s"
	policies "github.com/mansam/policy-runner/policies"
	"github.com/open-policy-agent/opa/rego"
)

const (
	Identifier = "bestpractices"
)

// New returns an Evaluator for the redhat-cop best practices policy set.
func New(policySet policies.PolicySet) (e *Evaluator) {
	filter := func(absPath string, info fs.FileInfo, depth int) bool {
		for _, exclude := range policySet.ExcludeSubstrings {
			if strings.Contains(absPath, exclude) {
				return true
			}
		}
		return false
	}
	e = &Evaluator{
		loader: rego.Load([]string{policySet.Path}, filter),
		query:  policySet.Query,
	}
	return
}

// Evaluator implementation for redhat-cop best practices policy set.
// https://github.com/redhat-cop/rego-policies/blob/main/policy/ocp/bestpractices
type Evaluator struct {
	loader func(r *rego.Rego)
	query  string
}

// Evaluate policies against the resource collection and report issues.
func (r *Evaluator) Evaluate(resources *k8s.UnstructuredResources) (issues []policies.Issue, err error) {
	query, err := rego.New(
		rego.Query(r.query),
		r.loader,
	).PrepareForEval(context.TODO())
	if err != nil {
		return
	}

	issueMap := make(map[string]*policies.Issue)
	for _, ns := range resources.NamespacedLists {
		for _, list := range ns {
			for _, item := range list.Items {
				resultSet, rErr := query.Eval(context.Background(), rego.EvalInput(item.UnstructuredContent()))
				if rErr != nil {
					err = rErr
				}
				for _, result := range resultSet {
					for _, exp := range result.Expressions {
						bytes, _ := json.Marshal(exp.Value)
						value := expressionValue{}
						_ = json.Unmarshal(bytes, &value)
						for k, v := range value {
							if issueMap[k] == nil && len(v.Violations) > 0 {
								issueMap[k] = &policies.Issue{
									Name:      k,
									ID:        v.Violations[0].Details.PolicyID,
									PolicySet: Identifier,
								}
							}
							for _, v := range v.Violations {
								issue := issueMap[k]
								incident := policies.Incident{
									Message:   v.Msg,
									Namespace: item.GetNamespace(),
									Name:      item.GetName(),
									Group:     item.GroupVersionKind().Group,
									Version:   item.GroupVersionKind().Version,
									Kind:      item.GroupVersionKind().Kind,
								}
								issue.Incidents = append(issue.Incidents, incident)
							}
						}
					}
				}
			}
		}
	}
	for _, p := range issueMap {
		issues = append(issues, *p)
	}
	return
}

// policy-specific rego output representations
type expressionValue map[string]bestPracticePolicy
type bestPracticePolicy struct {
	Violations []struct {
		Msg     string `json:"msg"`
		Details struct {
			PolicyID string `json:"policyID"`
		} `json:"details"`
	} `json:"violation"`
}
