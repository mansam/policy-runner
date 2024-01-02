package registry

import (
	"errors"

	"github.com/mansam/policy-runner/k8s"
	"github.com/mansam/policy-runner/policies"
	"github.com/mansam/policy-runner/policies/bestpractices"
)

// NewPolicyRegistry constructs a new policy registry.
func NewPolicyRegistry() (p *PolicyRegistry) {
	p = &PolicyRegistry{}
	p.Evaluators = make(map[string]policies.Evaluator)
	return
}

// PolicyRegistry stores known policy sets.
type PolicyRegistry struct {
	Evaluators map[string]policies.Evaluator
}

// EvaluateAll registered policy sets.
func (r *PolicyRegistry) EvaluateAll(ur *k8s.UnstructuredResources) (issues []policies.Issue, err error) {
	for _, e := range r.Evaluators {
		var i []policies.Issue
		i, err = e.Evaluate(ur)
		if err != nil {
			return
		}
		issues = append(issues, i...)
	}
	return
}

// RegisterPolicySet stores recognized policy sets in the registry.
func (r *PolicyRegistry) RegisterPolicySet(policySet policies.PolicySet) (err error) {
	switch policySet.Name {
	case bestpractices.Identifier:
		r.Evaluators[policySet.Name] = bestpractices.New(policySet)
	default:
		err = errors.New("Unrecognized policy set: " + policySet.Name)
	}
	return
}
