# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, please report it privately. **Do not** open public issues for security vulnerabilities.

### Contact

Email: security@aryeko.dev

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes (optional)

### Response Timeline

- **Acknowledgment:** within 72 hours
- **Initial assessment:** within 1 week
- **Fix timeline:** depends on severity

We will work with you on a fix and coordinate a responsible disclosure timeline.

## Security Advisories

Security fixes will be announced via [GitHub Security Advisories](https://github.com/go-modkit/modkit/security/advisories).

## Best Practices

When using modkit in production:

- Keep dependencies updated (`go get -u`)
- Run `govulncheck` regularly (included in `make vuln`)
- Follow the principle of least privilege for module exports
- Validate all user input in controllers before processing
