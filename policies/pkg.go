package policies

import "github.com/mansam/policy-runner/k8s"

type Evaluator interface {
	Evaluate(ur *k8s.UnstructuredResources) ([]Issue, error)
}

type PolicySet struct {
	Name              string
	Path              string
	ExcludeSubstrings []string
	Query             string
}

type Issue struct {
	PolicySet string     `json:"policySet"`
	Name      string     `json:"name"`
	Incidents []Incident `json:"incidents"`
	Severity  string     `json:"severity"`
	ID        string     `json:"id"`
}

type Incident struct {
	Message   string `json:"message"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
}
