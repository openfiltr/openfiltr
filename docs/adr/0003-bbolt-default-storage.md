# ADR-0003: bbolt as the Default Embedded Backend with PostgreSQL Compatibility

**Date:** 2026-03-30
**Status:** Accepted

## Context

ADR-0001 deliberately removed SQLite and standardised the codebase on PostgreSQL. That was useful while the storage layer, migrations, and query shapes were still being stabilised.

The merged bbolt migration in PR #126 changed the operating reality:

- the application now has a storage seam instead of hardwired `*sql.DB` coupling
- a deterministic bbolt store exists with bucket initialisation and metadata versioning
- auth, filtering, DNS entries, activity, stats, config import or export, and the remaining CRUD paths all run against bbolt
- the default installation and router deployment flows no longer need an external database service

The project still benefits from PostgreSQL compatibility for operators who want an external database, but the previous PostgreSQL-only ADR no longer describes the current runtime model.

## Decision

OpenFiltr now treats bbolt as the default runtime backend.

The application should:

- use `storage.database_path` as the normal default configuration path
- resolve relative database paths against the config directory
- start successfully without PostgreSQL when only a local bbolt file is configured
- continue to support PostgreSQL through `storage.database_url` as an optional backend

The codebase keeps the existing SQL and migration path for PostgreSQL compatibility, but product and deployment guidance should describe bbolt-first operation unless a document is explicitly about PostgreSQL.

## Consequences

**Positive:**
- Default installs become single-binary and self-contained.
- Router and other constrained deployments no longer need an external database service.
- Local development and smoke testing are simpler because a writable config directory is enough for the default path.
- PostgreSQL remains available for operators who prefer an external datastore.

**Negative:**
- The storage layer is now dual-path, so behaviour parity has to be maintained across bbolt and PostgreSQL.
- Documentation and agent guidance must be kept honest; stale PostgreSQL-only instructions become actively misleading.
- Some roadmap items still assume features are complete when the resolver path has not caught up, so documentation needs tighter maintenance.

**Neutral:**
- SQL migrations remain part of the repository because PostgreSQL support still exists.
- Historical ADR-0001 remains valuable as a record of why SQLite was removed even though its PostgreSQL-only conclusion has been superseded.

## References

- Supersedes [ADR-0001](0001-postgresql-only-storage.md)
- PR [#126](https://github.com/openfiltr/openfiltr/pull/126) — switch default storage to bbolt and add OpenWrt install path
- `internal/storage/store.go` — storage seam used by the runtime
- `internal/storage/bbolt_store.go` — bbolt bootstrap, bucket initialisation, and metadata versioning
- `internal/config/config.go` — default backend configuration
