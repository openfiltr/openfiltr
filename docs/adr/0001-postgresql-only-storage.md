# ADR-0001: PostgreSQL as the Only Supported Database Backend

**Date:** 2026-03-18
**Status:** Superseded by ADR-0003

## Context

The project was initially scaffolded with SQLite as its embedded storage backend. SQLite is convenient for local development but introduces significant constraints for a self-hosted, multi-tenant DNS filtering platform:

- SQLite's WAL mode limits concurrent write throughput.
- Deploying alongside a container orchestrator (Docker Compose, Kubernetes) is simpler with a dedicated database service.
- SQLite's SQL dialect diverges from PostgreSQL in ways that create ongoing maintenance friction (placeholder syntax, date functions, upsert syntax).
- The target deployment audience (self-hosters) typically already runs PostgreSQL or can run it via Docker Compose.

## Decision

Remove SQLite entirely. PostgreSQL (13 or later) is the only supported database backend. The application connects via a DSN configured through `storage.database_url` in `config/app.yaml` or the `OPENFILTR_DATABASE_URL` environment variable.

All SQL is written in PostgreSQL dialect:
- Positional placeholders (`$1`, `$2`, …) not `?`.
- `NOW()` not `datetime('now')`.
- `ON CONFLICT … DO UPDATE` for upserts.

Schema is managed through versioned SQL migration files under `internal/storage/migrations/`.

## Consequences

**Positive:**
- Concurrent reads and writes are handled robustly.
- Richer query capabilities (window functions, CTEs, `RETURNING`).
- Cleaner separation between application and storage processes.
- No CGO dependency; the `pgx` or `lib/pq` drivers are pure Go.

**Negative:**
- Local development now requires a running PostgreSQL instance (mitigated by the provided Docker Compose file).
- A `docker-compose.yml` with a `postgres` service is mandatory for contributors.

**Neutral:**
- All existing SQL call sites were ported in the same change set; no feature behaviour changed.

## References

- Issue: migration from SQLite to PostgreSQL (internal planning)
- PR implementing the change: merged 2026-03-18
- `internal/storage/migrations/001_initial_schema.sql` — PostgreSQL-native initial schema
- `deploy/docker/docker-compose.yml` — reference Compose file with PostgreSQL service
- Superseded by [ADR-0003](0003-bbolt-default-storage.md)
