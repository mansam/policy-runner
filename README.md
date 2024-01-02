# policy-runner

1. Copy config.json.example and set the desired namespaces and kubeconfig path.
2. `make cli`
3. `make checkout-policies` to checkout the redhat-cop best practices rego policy set.
4. Run `./bin/policy-runner -config path-to-config.json`. If you want to run this from
   anywhere other than the root directory, you will need to adjust the path to the
   policy set in the config file.
