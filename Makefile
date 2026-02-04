SHELL := /bin/sh

.PHONY: fmt lint vuln test

GOPATH ?= $(shell go env GOPATH)
GOIMPORTS ?= $(GOPATH)/bin/goimports
GOLANGCI_LINT ?= $(GOPATH)/bin/golangci-lint
GOVULNCHECK ?= $(GOPATH)/bin/govulncheck

fmt:
	gofmt -w .
	$(GOIMPORTS) -w .

lint:
	$(GOLANGCI_LINT) run

vuln:
	$(GOVULNCHECK) ./...

test:
	go test ./...
