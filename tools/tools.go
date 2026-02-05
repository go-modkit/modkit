//go:build tools

// Package tools tracks development tool dependencies.
package tools

import (
	_ "github.com/conventionalcommit/commitlint"
	_ "github.com/evilmartians/lefthook/v2"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/seriousben/go-patch-cover/cmd/go-patch-cover"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
