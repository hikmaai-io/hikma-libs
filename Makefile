# ABOUTME: Build targets for hikma-libs shared Go libraries.
# ABOUTME: Each subdirectory is a separate Go module.

.PHONY: test fmt vet lint vulncheck check

## Run tests with race detector
test:
	cd hlog && go test -race -count=1 ./...

## Format code
fmt:
	gofmt -s -w .

## Static analysis
vet:
	cd hlog && go vet ./...

## Run golangci-lint
lint:
	cd hlog && golangci-lint run ./...

## Check for known vulnerabilities
vulncheck:
	cd hlog && $(shell go env GOPATH)/bin/govulncheck ./...

## Full check: fmt + vet + vulncheck + test
check: fmt vet vulncheck test
