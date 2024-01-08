package policy_provider

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-logr/logr"
	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/mansam/policy-runner/k8s"
	"github.com/mansam/policy-runner/policies"
	"github.com/mansam/policy-runner/policies/registry"
	"go.lsp.dev/uri"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CLIConfig struct {
	Namespaces        []string                  `json:"namespaces"`
	GroupVersionKinds []schema.GroupVersionKind `json:"groupVersionKinds"`
	PolicySets        []policies.PolicySet      `json:"policySets"`
	KubeConfigPath    string                    `json:"kubeConfigPath"`
}

type policyProvider struct {
	client    client.Client
	cliConfig CLIConfig
	registry  registry.PolicyRegistry
	provider.UnimplementedDependenciesComponent
}

// Evaluate implements provider.ServiceClient.
func (p *policyProvider) Evaluate(ctx context.Context, cap string, conditionInfo []byte) (provider.ProviderEvaluateResponse, error) {
	resources := k8s.NewUnstructuredResources(p.client)
	for _, ns := range p.cliConfig.Namespaces {
		err := resources.Gather(ns, p.cliConfig.GroupVersionKinds)
		if err != nil {
			panic(err)
		}
	}

	issues, err := p.registry.EvaluateAll(resources)
	if err != nil {
		panic(err)
	}
	response := provider.ProviderEvaluateResponse{}

	if len(issues) > 0 {
		return response, nil
	}

	incidents := []provider.IncidentContext{}
	for _, i := range issues {
		for _, c := range i.Incidents {
			m, err := p.client.RESTMapper().RESTMapping(schema.GroupKind{Group: c.Group, Kind: c.Kind}, c.Version)
			if err != nil {
				return response, err
			}
			fileURI := "https://cluster-uri-api"
			if m.GroupVersionKind.Group == "" {
				fileURI = fmt.Sprintf("%v/%v/%v", fileURI, "api", m.Resource.Version)
			} else {
				fileURI = fmt.Sprintf("%v/%v/%v/%v", fileURI, "apis", m.Resource.Group, m.Resource.Version)
			}
			if c.Namespace == "" {
				fileURI = fmt.Sprintf("%v/%v/%v", fileURI, "namespaces", c.Namespace)
			}
			fileURI = fmt.Sprintf("%v/%v/%v", fileURI, m.Resource.Resource, c.Name)

			incidents = append(incidents, provider.IncidentContext{
				FileURI: uri.URI(fileURI),
				Variables: map[string]interface{}{
					"message": c.Message,
				},
				IsDependencyIncident: false,
			})
		}

	}
	response.Incidents = incidents
	return response, nil
}

// Stop implements provider.ServiceClient.
func (p *policyProvider) Stop() {
}

// Capabilities implements provider.BaseClient.
func (p *policyProvider) Capabilities() []provider.Capability {
	return []provider.Capability{
		{
			Name:            "rego",
			TemplateContext: openapi3.SchemaRef{},
		},
	}
}

// Init implements provider.BaseClient.
func (p *policyProvider) Init(context.Context, logr.Logger, provider.InitConfig) (provider.ServiceClient, error) {
	return p, nil

}

func NewPolicyProvider(k8sClient client.Client, cliConfig CLIConfig) (provider.BaseClient, error) {
	// in the future, this should come for the rule through the condition.
	policyRegistry := registry.NewPolicyRegistry()
	for _, policySet := range cliConfig.PolicySets {
		err := policyRegistry.RegisterPolicySet(policySet)
		if err != nil {
			panic(err)
		}
	}
	return &policyProvider{
		client:    k8sClient,
		cliConfig: cliConfig,
		registry:  *policyRegistry,
	}, nil
}
