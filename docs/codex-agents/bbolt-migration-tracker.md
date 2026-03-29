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
| 1 | [#115](https://github.com/openfiltr/openfiltr/issues/115) | Storage seam refactor: replace direct `*sql.DB` coupling with repository interface | Todo | _TBD_ | No behaviour change allowed. |
| 2 | [#116](https://github.com/openfiltr/openfiltr/issues/116) | Add bbolt store bootstrap, bucket initialisation, and metadata versioning | Todo | _TBD_ | Deterministic bucket creation and store version metadata. |
| 3 | [#117](https://github.com/openfiltr/openfiltr/issues/117) | Config and startup backend selector: support `storage.database_path` for bbolt | Todo | _TBD_ | Startup must succeed without PostgreSQL when path is configured. |
| 4 | [#118](https://github.com/openfiltr/openfiltr/issues/118) | Port auth persistence (`users` and `api_tokens`) to bbolt-backed repository | Todo | _TBD_ | Preserve JWT and CSRF behaviour. |
| 5 | [#119](https://github.com/openfiltr/openfiltr/issues/119) | Port filtering rules and DNS entries to bbolt with secondary indexes | Todo | _TBD_ | Preserve exact, wildcard, and regex matching semantics. |
| 6 | [#120](https://github.com/openfiltr/openfiltr/issues/120) | Port activity log and stats queries to bbolt | Todo | _TBD_ | Keep bounded runtime memory and write overhead. |
| 7 | [#121](https://github.com/openfiltr/openfiltr/issues/121) | Port remaining API CRUD resources and config import or export to bbolt | Todo | _TBD_ | Preserve API payload contracts. |
| 8 | [#122](https://github.com/openfiltr/openfiltr/issues/122) | Make bbolt the default backend and remove mandatory PostgreSQL startup dependency | Todo | _TBD_ | PostgreSQL optional legacy support only if trivial. |
| 9 | [#123](https://github.com/openfiltr/openfiltr/issues/123) | Document OpenWrt MT3000 single-binary deployment with bbolt and procd | Todo | _TBD_ | Include procd example and dnsmasq port guidance. |

## Changelog policy for this migration

Every PR that changes code, config, docs, or workflows for the migration must include an `Unreleased` entry in `CHANGELOG.md`.

When updating the changelog:

1. Add entries only under `## [Unreleased]`.
2. Use Keep a Changelog categories.
3. Include the pull request number and merge-relevant context.
4. Confirm the wording was checked by AI before merge and note that in the PR description.
