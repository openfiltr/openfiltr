# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| latest  | ✅        |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report vulnerabilities privately via [GitHub Security Advisories](https://github.com/openfiltr/openfiltr/security/advisories/new).

### Response timeline

| Stage | Target |
|-------|--------|
| Acknowledgement | 48 hours |
| Initial triage | 5 business days |
| Fix / mitigation | 90 days |

## Security baseline

OpenFiltr ships with:

- CSRF protection for browser sessions
- Secure, `HttpOnly` session cookies
- Rate limiting on authentication endpoints
- Audit logging for all destructive actions
- Signed releases (planned for v1.0)
- SBOM generation on every release
