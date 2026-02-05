SHELL := /bin/sh

.PHONY: fmt lint vuln test test-coverage test-patch-coverage tools setup-hooks lint-commit

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

test-patch-coverage:
	@changed=$$(git diff --name-only origin/main...HEAD --diff-filter=AM | grep '\.go$$' || true); \
	if [ -z "$$changed" ]; then \
		echo "No changed Go files detected vs origin/main."; \
		exit 0; \
	fi; \
	pkgs=$$(echo "$$changed" | xargs -n1 dirname | sort -u | awk '{ if ($$0 == ".") { print "./..." } else { print "./"$$0"/..." } }' | tr '\n' ' '); \
	echo "Running patch coverage for packages:"; \
	echo "$$pkgs" | tr ' ' '\n'; \
	go test -race -coverprofile=coverage-patch.out -covermode=atomic $$pkgs; \
	exclude_pattern=$$(awk '/^ignore:/{flag=1; next} /^[a-z]/ && !/^ignore:/{flag=0} flag && /^  -/{gsub(/^[[:space:]]*-[[:space:]]*"/,""); gsub(/"$$/,""); print}' codecov.yml 2>/dev/null | tr '\n' '|' | sed 's/|$$//'); \
	if [ -n "$$exclude_pattern" ]; then \
		grep -vE "$$exclude_pattern" coverage-patch.out > coverage-patch-filtered.out || cp coverage-patch.out coverage-patch-filtered.out; \
		go tool cover -func=coverage-patch-filtered.out; \
	else \
		go tool cover -func=coverage-patch.out; \
	fi

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
