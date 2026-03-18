## Repository Truth

- This repository is currently a backend-first Go service.
- Treat the API, DNS, auth, config, storage, Docker, and workflow files as the source of truth.
- Do not assume a working React UI or SDK exists just because older docs mention them.

## Working Rules

- Use British English in code comments, docs, and commit messages.
- Be direct. Do not pad explanations or hide obvious risks.
- Prefer small, scoped changes over wide speculative refactors.
- Use `rg` for search.
- Sign commits with `-s`.
- Branch names should use the `codex/` prefix.

## Repo Map

- `cmd/server/main.go`: process bootstrap
- `internal/config`: config loading and defaults
- `internal/storage`: database bootstrap, migrations, SQL helpers
- `internal/auth`: password hashing, JWTs, API token validation
- `internal/api`: HTTP handlers and routing
- `internal/dns`: DNS request handling and activity logging
- `deploy/docker`: container image and compose examples
- `.github/workflows`: CI, release, CodeQL, dependency review

## Storage Notes

- PostgreSQL is the only supported database backend.
- Configuration should use `storage.database_url` or `OPENFILTR_DATABASE_URL`.
- Do not reintroduce SQLite-specific code, docs, or workflow dependencies.

## Worktrees

- Local isolated work should go under `.worktrees/`.
- `.worktrees/` is intentionally ignored in `.gitignore`.
- Use a worktree when starting non-trivial feature work off `origin/main` so you do not contaminate another branch.

## Verification

- If `go` is unavailable locally, use Docker for Go tooling:
  - `docker run --rm -v "$PWD":/work -w /work golang:1.24 go test ./...`
  - `docker run --rm -v "$PWD":/work -w /work golang:1.24 gofmt -w <files>`
- Parse YAML after workflow or config edits.
- Grep for stale SQLite references after storage changes.

## Current Risks

- README and roadmap history may lag behind the code.
- Storage changes touch many raw SQL call sites.
- CI and release workflows are part of the product surface here, not afterthoughts.
