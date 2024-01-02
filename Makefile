cli: fmt vet
	go build -ldflags="-w -s" -o bin/policy-runner cmd/cli/main.go

fmt:
	go fmt ./...

vet:
	go vet ./...

checkout-policies:
	git clone git@github.com:redhat-cop/rego-policies.git