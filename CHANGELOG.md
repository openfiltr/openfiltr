# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Architecture Decision Records (ADRs) directory at `docs/adr/` with a README template and an index. Two initial ADRs document the PostgreSQL-only storage decision (ADR-0001) and the agent pre-task protocol (ADR-0002).
- Automated changelog update workflow (`.github/workflows/changelog-update.yml`): on every PR merge to `main`, a structured GitHub issue is created and assigned to the Copilot coding agent, prompting it to append an entry to `CHANGELOG.md` with full PR context (title, author, merge date, description).

### Changed

- `AGENTS.md` and `.github/copilot-instructions/project.md` updated with a mandatory pre-task protocol: AI coding agents must sync with `main`, read `CHANGELOG.md`, read all ADRs, and run `go test ./...` before making any code changes.
- Replaced SQLite storage with PostgreSQL-backed configuration and runtime support.

### Fixed

- PR #93, merged on 2026-03-21, added exact, wildcard, apex-domain, and regex block rule matching in the DNS server, normalised queried domains for lookup, and documented the Codex branch and PR workflow so future agents follow the repository process.

## [0.1.0] — 2026-03-18

### Added

- Initial project scaffolding: Go backend, React frontend, community files.
- REST API covering filtering, DNS, clients, activity, config, and auth.
- DNS filtering server with block/allow rules and upstream forwarding.
- SQLite storage with embedded migrations.
- JWT and API token authentication.
- Docker and curl install support.
- OpenAPI 3.1 specification.
- GitHub Actions CI, release, and security workflows.
- GitHub issue infrastructure (labels, issues, create-issues workflow).

> This project was bootstrapped with the assistance of AI. See CONTRIBUTING.md.
