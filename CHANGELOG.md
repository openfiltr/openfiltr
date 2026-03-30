# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- In-repo bbolt migration tracker added at `docs/codex-agents/bbolt-migration-tracker.md` with ordered issue mapping (#115 to #123), status guidance, and a requirement to record implementation PR links for future agents.
- New pull request guard workflow `.github/workflows/changelog-guard.yml` now requires `CHANGELOG.md` updates on non-trivial PRs and validates that added changelog text includes at least one bullet under `## [Unreleased]`.
- Architecture Decision Records (ADRs) directory at `docs/adr/` with a README template and an index. Two initial ADRs document the PostgreSQL-only storage decision (ADR-0001) and the agent pre-task protocol (ADR-0002).
- Automated changelog update workflow (`.github/workflows/changelog-update.yml`): on every PR merge to `main`, a structured GitHub issue is created and assigned to the Copilot coding agent, prompting it to append an entry to `CHANGELOG.md` with full PR context (title, author, merge date, description).
- Added a BusyBox-friendly OpenWrt installer at `scripts/install-openwrt.sh` for MT3000 and MT6000 routers. It downloads router release assets, prompts for router IP and ports, installs the binary plus `procd` service, and configures dnsmasq forwarding or exclusive port 53 mode.
- PR #124 (merged 2026-03-29): added `docs/codex-agents/README.md` guidance requiring changelog updates for implementation PRs and instructing agents to update the bbolt migration tracker where applicable, ensuring both human and AI contributors follow the pre-task and post-task protocols.

### Changed

- Added a storage interface seam across startup, API, DNS, and auth so the code no longer depends on concrete `*sql.DB` call sites for the bbolt migration work (#115).
- Added an isolated bbolt bootstrap store with deterministic bucket initialisation, metadata versioning, and reopen tests in `internal/storage/bbolt_store.go` (#116).
- Added `storage.database_path` plus startup backend selection so the server can boot against bbolt without requiring PostgreSQL at launch (#117).
- Added an auth repository layer for `users` and `api_tokens` so login, setup, API token management, and API token validation work against bbolt as well as PostgreSQL (#118).
- Added bbolt lookup buckets for block rules, allow rules, and DNS entries so filtering and DNS lookups use secondary indexes instead of SQL-only queries (#119).
- Added bbolt activity-log writes plus in-memory count, list, and top-domain queries so system stats and activity endpoints work without PostgreSQL (#120).
- Added bbolt-backed CRUD for the remaining API resources plus config export/import support, keeping the payload shape aligned with the PostgreSQL path (#121).
- Made bbolt the default runtime backend, resolved relative database paths against the config directory, and dropped the mandatory PostgreSQL startup dependency (#122).
- Added an OpenWrt MT3000 deployment guide with `procd` service wiring and dnsmasq port guidance (#123).
- CI now builds OpenWrt arm64 router archives, and tagged releases now publish `openfiltr-openwrt-arm64.tar.gz`, `openfiltr-openwrt-mt3000.tar.gz`, and `openfiltr-openwrt-mt6000.tar.gz` alongside the existing Docker image and desktop release assets.
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
