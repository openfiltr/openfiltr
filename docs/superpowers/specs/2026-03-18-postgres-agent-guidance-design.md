# PostgreSQL Storage And Agent Guidance Design

**Date:** 2026-03-18

## Goal

Replace SQLite with PostgreSQL as the only supported database backend and add one root `AGENTS.md` file that gives coding agents an accurate, concise map of the repository, coding style, workflow expectations, and current project caveats.

## Why

SQLite is currently baked into the runtime, schema, configuration, CI, Docker setup, and documentation. That is a poor fit for:

- production deployment
- higher write concurrency and more predictable reliability
- future features that will add more state, background work, and cross-entity operations

The repo also has no local agent instructions, which means code agents have to infer too much from a codebase whose README currently overstates what is implemented.

## Decisions

### Database

- PostgreSQL replaces SQLite outright.
- There is no compatibility layer and no data migration path.
- Database configuration becomes PostgreSQL-first instead of file-path based.
- Migrations move out of embedded Go strings and into versioned SQL files.
- The codebase should use PostgreSQL-compatible SQL and placeholder syntax everywhere.

### Scope boundaries

Included:

- runtime database bootstrap
- configuration model and environment variables
- schema and migration mechanism
- SQL query compatibility fixes
- local Docker/development setup
- CI and release workflow updates that mention SQLite
- documentation updates that currently claim SQLite
- one root `AGENTS.md`

Excluded:

- SQLite to PostgreSQL data migration
- introducing a repository layer or query generation framework
- adding a second supported database backend
- wider architectural clean-up unrelated to storage or agent guidance

## Target architecture

### Storage bootstrap

`internal/storage` remains the entry point for opening the database and running migrations, but it becomes PostgreSQL-specific instead of pretending to be generic.

Expected shape:

- `internal/storage/db.go` opens a PostgreSQL connection pool
- `internal/storage/migrations.go` applies SQL migrations from a migration directory
- `internal/storage/migrations/*.sql` contains ordered schema files

### Configuration

The current `storage.path` field is removed. Replace it with explicit PostgreSQL settings or a DSN. Prefer a DSN because it is simpler for Docker, CI, and deployment.

Expected sources:

- YAML config example for local/development defaults
- `OPENFILTR_DATABASE_URL` environment override

### Schema

The schema should stop carrying SQLite baggage where it is obvious.

Expected PostgreSQL choices:

- `UUID` for ids
- `BOOLEAN` instead of integer flags
- `TIMESTAMPTZ` instead of `DATETIME`
- `JSONB` for structured JSON text fields if those fields are actually treated as JSON

The schema can stay close to the current tables to reduce application churn.

### Application SQL

All application queries need to move from SQLite assumptions to PostgreSQL assumptions:

- `?` placeholders become `$1`, `$2`, ...
- `datetime('now')` becomes `NOW()`
- string concatenation and wildcard logic must remain valid in PostgreSQL
- `ON CONFLICT` clauses need PostgreSQL syntax
- pagination queries stay structurally the same

### Agent guidance

Add a root `AGENTS.md` that covers:

- what the repo actually contains today
- British English requirement
- branch naming and commit expectations
- preferred commands for search, build, tests, and workflows
- where storage, DNS, auth, API, config, and workflow code live
- warnings about README drift and absent UI/SDK code
- worktree guidance, including using local `.worktrees/` safely

## File map

Likely new files:

- `AGENTS.md`
- `docs/superpowers/plans/2026-03-18-postgres-agent-guidance.md`
- `internal/storage/migrations/001_initial_schema.sql`

Likely modified files:

- `.gitignore`
- `README.md`
- `ROADMAP.md`
- `CONTRIBUTING.md`
- `config/app.yaml.example`
- `deploy/docker/Dockerfile`
- `deploy/docker/docker-compose.yml`
- `examples/docker-compose.yml`
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `cmd/server/main.go`
- `internal/config/config.go`
- `internal/storage/db.go`
- `internal/storage/migrations.go`
- `internal/auth/auth.go`
- `internal/api/*.go`
- `internal/dns/server.go`

## Risks

- SQL compatibility churn is spread across many handlers, so careless edits will break basic CRUD paths.
- PostgreSQL schema choices made now will influence later auth, audit, and policy features.
- Some docs currently describe a more complete product than the tree actually contains; agent guidance must correct that or future automated changes will drift further.
- Local verification is limited in this environment because the Go toolchain is not installed.

## Verification strategy

- parse YAML files after workflow/config edits
- grep for leftover SQLite driver references and `?` placeholders in SQL
- grep for leftover SQLite-specific functions such as `datetime('now')`
- where possible, run repository checks in CI after push

