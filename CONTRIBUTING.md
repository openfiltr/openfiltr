# Contributing to OpenFiltr

Thank you for your interest in contributing to OpenFiltr. This project was bootstrapped with the assistance of AI — see the note in the README for context.

## Developer Certificate of Origin

All commits must be signed off under the [Developer Certificate of Origin](https://developercertificate.org). Add `-s` to your commit command:

```bash
git commit -s -m "feat(dns): add wildcard rule support"
```

## Prerequisites

- Go 1.24+
- Docker (optional, for integration tests)
- `make`

## Local setup

```bash
git clone https://github.com/openfiltr/openfiltr
cd openfiltr
go mod download
make build        # compile the server
make test         # run all tests
```

## Commit message format

```
type(scope): short description in British English

Types: feat, fix, docs, refactor, test, chore
Scopes: dns, api, auth, config, frontend, ci, docs
```

## Submitting a pull request

1. Fork the repository and create a feature branch.
2. Make your changes with tests.
3. Run `make test` and `make lint`.
4. Open a pull request against `main` using the PR template.

## Code of Conduct

This project follows the [Contributor Covenant 2.1](CODE_OF_CONDUCT.md).
