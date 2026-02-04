# Tooling

This repository uses standard Go tools plus common OSS linters.

## Format

```bash
make fmt
```

Runs:
- `gofmt -w .`
- `goimports -w .`

Install:

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

## Lint

```bash
make lint
```

Runs:
- `golangci-lint run`

See `.golangci.yml` for enabled linters and excluded paths.

Install:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Vulnerability Scan

```bash
make vuln
```

Runs:
- `govulncheck ./...`

Install:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

## Test

```bash
make test
```

Runs:
- `go test ./...`
