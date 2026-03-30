# bbolt Migration Tracker

This file is the in-repo execution board for the bbolt migration programme.

All AI and human contributors should update this file in the same pull request as the implementation change they ship.

## How to use this tracker

- Keep issue scope small and single-purpose.
- Set `Status` to one of: `Todo`, `In progress`, `Blocked`, `Done`.
- Record the implementation PR once work lands.
- Record any behaviour notes relevant to future agents.
- Do not delete historical rows. Mark superseded work clearly instead.

## Work queue

| Order | Issue | Title | Status | Implementation PR | Notes |
|------:|:------|:------|:-------|:------------------|:------|
| 1 | [#115](https://github.com/openfiltr/openfiltr/issues/115) | Storage seam refactor: replace direct `*sql.DB` coupling with repository interface | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | Introduced the storage seam used by both the bbolt and PostgreSQL paths. |
| 2 | [#116](https://github.com/openfiltr/openfiltr/issues/116) | Add bbolt store bootstrap, bucket initialisation, and metadata versioning | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | Deterministic bucket creation, store version metadata, and isolated reopen tests live in `internal/storage/bbolt_store.go`. |
| 3 | [#117](https://github.com/openfiltr/openfiltr/issues/117) | Config and startup backend selector: support `storage.database_path` for bbolt | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | Startup now selects bbolt when `storage.database_path` is set and no longer requires PostgreSQL for the default path. |
| 4 | [#118](https://github.com/openfiltr/openfiltr/issues/118) | Port auth persistence (`users` and `api_tokens`) to bbolt-backed repository | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | Login, setup, token listing, token creation, token deletion, and API token validation now use the auth repository layer for both PostgreSQL and bbolt. |
| 5 | [#119](https://github.com/openfiltr/openfiltr/issues/119) | Port filtering rules and DNS entries to bbolt with secondary indexes | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | bbolt now maintains lookup buckets for enabled rules and DNS entries, and the DNS server uses them for exact, wildcard, regex, and host or type lookups. |
| 6 | [#120](https://github.com/openfiltr/openfiltr/issues/120) | Port activity log and stats queries to bbolt | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | bbolt activity logs now back system stats, activity listings, top-domain aggregation, and DNS query writes. |
| 7 | [#121](https://github.com/openfiltr/openfiltr/issues/121) | Port remaining API CRUD resources and config import or export to bbolt | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | Config export or import now runs against bbolt, and audit event listing uses the bbolt store directly while preserving API payload contracts. |
| 8 | [#122](https://github.com/openfiltr/openfiltr/issues/122) | Make bbolt the default backend and remove mandatory PostgreSQL startup dependency | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | Default config now uses a local bbolt database beside the config file; PostgreSQL is only used when `storage.database_url` is set. |
| 9 | [#123](https://github.com/openfiltr/openfiltr/issues/123) | Document OpenWrt MT3000 single-binary deployment with bbolt and procd | Done | [#126](https://github.com/openfiltr/openfiltr/pull/126) | Added a dedicated OpenWrt MT3000 and MT6000 guide with a `procd` service example and dnsmasq port guidance. |

## Changelog policy for this migration

Every PR that changes code, config, docs, or workflows for the migration must include an `Unreleased` entry in `CHANGELOG.md`.

When updating the changelog:

1. Add entries only under `## [Unreleased]`.
2. Use Keep a Changelog categories.
3. Include the pull request number and merge-relevant context.
4. Confirm the wording was checked by AI before merge and note that in the PR description.
