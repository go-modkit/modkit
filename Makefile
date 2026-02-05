SHELL := /bin/sh

.PHONY: fmt lint vuln test test-coverage tools setup-hooks lint-commit

GOPATH ?= $(shell go env GOPATH)
GOIMPORTS ?= $(GOPATH)/bin/goimports
GOLANGCI_LINT ?= $(GOPATH)/bin/golangci-lint
GOVULNCHECK ?= $(GOPATH)/bin/govulncheck
LEFTHOOK ?= $(GOPATH)/bin/lefthook
COMMITLINT ?= $(GOPATH)/bin/commitlint

fmt: tools
	gofmt -w .
	$(GOIMPORTS) -w .

lint:
	$(GOLANGCI_LINT) run

vuln:
	$(GOVULNCHECK) ./...

test:
	go test ./...

test-coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go test -race -coverprofile=coverage-examples.out -covermode=atomic ./examples/hello-mysql/...

# Install all development tools (tracked in tools/tools.go)
tools:
	@echo "Installing development tools..."
	@cat tools/tools.go | grep _ | awk '{print $$2}' | xargs -I {} sh -c 'go install {}'
	@echo "Done: All tools installed"

# Install development tools and setup git hooks
setup-hooks: tools
	@echo "Setting up git hooks..."
	@if ! command -v lefthook >/dev/null 2>&1; then \
		echo "Warning: lefthook not found in PATH. Ensure \$$GOPATH/bin is in your PATH:"; \
		echo "  export PATH=\"\$$(go env GOPATH)/bin:\$$PATH\""; \
	fi
	$(LEFTHOOK) install
	@echo "Done: Git hooks installed"

# Validate a commit message (for manual testing)
lint-commit:
	@echo "$(MSG)" | $(COMMITLINT) lint
