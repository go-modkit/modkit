SHELL := /bin/sh

.PHONY: fmt lint vuln test test-coverage test-patch-coverage tools setup-hooks lint-commit

GOPATH ?= $(shell go env GOPATH)
GOIMPORTS ?= $(GOPATH)/bin/goimports
GOLANGCI_LINT ?= $(GOPATH)/bin/golangci-lint
GOVULNCHECK ?= $(GOPATH)/bin/govulncheck
GO_PATCH_COVER ?= $(GOPATH)/bin/go-patch-cover
LEFTHOOK ?= $(GOPATH)/bin/lefthook
COMMITLINT ?= $(GOPATH)/bin/commitlint

# Find all directories with go.mod, excluding hidden dirs (like .worktrees) and vendor
MODULES = $(shell find . -type f -name "go.mod" -not -path "*/.*/*" -not -path "*/vendor/*" -exec dirname {} \;)

fmt: tools
	gofmt -w .
	$(GOIMPORTS) -w .

lint:
	$(GOLANGCI_LINT) run

vuln:
	$(GOVULNCHECK) ./...

test:
	@for mod in $(MODULES); do \
		echo "Testing module: $$mod"; \
		(cd $$mod && go test -race -timeout=5m ./...) || exit 1; \
	done

test-coverage:
	@mkdir -p .coverage
	@echo "mode: atomic" > .coverage/coverage.out
	@for mod in $(MODULES); do \
		echo "Testing coverage for module: $$mod"; \
		(cd $$mod && go test -race -coverprofile=profile.out -covermode=atomic ./... || exit 1); \
		if [ -f $$mod/profile.out ]; then \
			tail -n +2 $$mod/profile.out >> .coverage/coverage.out; \
			rm $$mod/profile.out; \
		fi; \
	done
	@echo "\nTotal Coverage:"
	@go tool cover -func=.coverage/coverage.out | grep "total:"

test-patch-coverage: test-coverage
	@echo "Comparing against origin/main..."
	@git diff -U0 --no-color origin/main...HEAD > .coverage/diff.patch
	@$(GO_PATCH_COVER) .coverage/coverage.out .coverage/diff.patch > .coverage/patch_coverage.out
	@echo "Patch Coverage Report:"
	@cat .coverage/patch_coverage.out

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
