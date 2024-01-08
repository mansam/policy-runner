package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bombsimon/logrusr/v3"
	"github.com/konveyor/analyzer-lsp/provider"
	"github.com/mansam/policy-runner/k8s"
	"github.com/mansam/policy-runner/policy_provider"
	"github.com/sirupsen/logrus"
)

// Flags
var (
	ConfigPath string
	Port       int
)

func init() {
	flag.StringVar(&ConfigPath, "config", "config.json", "Path to config file (json)")
	flag.IntVar(&Port, "port", 0, "Port must be set")
}

func main() {
	flag.Parse()

	cliConfigBytes, err := os.ReadFile(ConfigPath)
	if err != nil {
		panic(err)
	}
	cliConfig := policy_provider.CLIConfig{}
	err = json.Unmarshal(cliConfigBytes, &cliConfig)
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

	logrusLog := logrus.New()
	logrusLog.SetOutput(os.Stdout)
	logrusLog.SetFormatter(&logrus.TextFormatter{})
	// need to do research on mapping in logrusr to level here TODO
	logrusLog.SetLevel(logrus.Level(5))

	log := logrusr.New(logrusLog)

	client, err := policy_provider.NewPolicyProvider(k8sClient, cliConfig)
	if err != nil {
		panic(err)
	}

	if Port == 0 {
		panic(fmt.Errorf("must pass in the port for the external provider"))
	}

	s := provider.NewServer(client, Port, log)
	ctx := context.TODO()
	s.Start(ctx)

	resources := k8s.NewUnstructuredResources(k8sClient)
	for _, ns := range cliConfig.Namespaces {
		err = resources.Gather(ns, cliConfig.GroupVersionKinds)
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
