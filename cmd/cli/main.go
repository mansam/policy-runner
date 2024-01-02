package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/mansam/policy-runner/k8s"
	"github.com/mansam/policy-runner/policies"
	"github.com/mansam/policy-runner/policies/registry"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Flags
var (
	ConfigPath string
)

type CLIConfig struct {
	Namespaces        []string                  `json:"namespaces"`
	GroupVersionKinds []schema.GroupVersionKind `json:"groupVersionKinds"`
	PolicySets        []policies.PolicySet      `json:"policySets"`
	KubeConfigPath    string                    `json:"kubeConfigPath"`
}

func init() {
	flag.StringVar(&ConfigPath, "config", "config.json", "Path to config file (json)")
}

func main() {
	flag.Parse()

	cliConfigBytes, err := os.ReadFile(ConfigPath)
	if err != nil {
		panic(err)
	}
	cliConfig := &CLIConfig{}
	err = json.Unmarshal(cliConfigBytes, cliConfig)
	if err != nil {
		panic(err)
	}
	kubeConfig, err := os.ReadFile(cliConfig.KubeConfigPath)
	if err != nil {
		panic(err)
	}
	k8sClient, err := k8s.NewClient(kubeConfig)
	if err != nil {
		panic(err)
	}

	resources := k8s.NewUnstructuredResources(k8sClient)
	for _, ns := range cliConfig.Namespaces {
		err = resources.Gather(ns, cliConfig.GroupVersionKinds)
		if err != nil {
			panic(err)
		}
	}

	policyRegistry := registry.NewPolicyRegistry()
	for _, policySet := range cliConfig.PolicySets {
		err = policyRegistry.RegisterPolicySet(policySet)
		if err != nil {
			panic(err)
		}
	}

	issues, err := policyRegistry.EvaluateAll(resources)
	if err != nil {
		panic(err)
	}
	bytes, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", bytes)
}
