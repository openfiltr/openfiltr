## Before You Begin Any Task

Complete these steps **before making any code changes**:

1. **Sync with `main`** — fetch and rebase onto the latest default branch:
   ```bash
   git fetch origin main
   git rebase origin/main
   ```

2. **Read `CHANGELOG.md`** — understand what has recently changed and why.

3. **Read all ADRs in `docs/adr/`** — understand the architectural decisions that have been deliberately made so you do not accidentally reverse them.

4. **Run the test suite** to record the current passing baseline before you touch anything:
   ```bash
   go test ./...
   # or, if Go is not available locally:
   docker run --rm -v "$PWD":/work -w /work golang:1.24 go test ./...
   ```

> Skipping any of these steps risks working from stale state, reversing deliberate decisions, or masking pre-existing test failures as regressions you introduced.

---

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
- `docs/adr/`: Architecture Decision Records (read before starting any task)

## Storage Notes

- PostgreSQL is the only supported database backend.
- Configuration should use `storage.database_url` or `OPENFILTR_DATABASE_URL`.
- Do not reintroduce SQLite-specific code, docs, or workflow dependencies.

## Worktrees

- Local isolated work should go under `.worktrees/`.
- `.worktrees/` is intentionally ignored in `.gitignore`.
- Use a worktree when starting non-trivial feature work off `origin/main` so you do not contaminate another branch.

## Branching Strategy

- Start non-trivial work from the latest `origin/main`, not from an old feature branch.
- Create one scoped branch per concern under the `codex/` prefix.
- Open PRs into `main`; do not stack unrelated work onto an already-open feature PR.
- If a PR branch drifts, merge or rebase the latest `main` into that branch before adding more work.
- Keep temporary rescue branches exceptional. If branch rules block direct updates, use a replacement branch only long enough to unblock the original PR or replace it cleanly.

## Merge Gates

- `main` is protected by a repository ruleset. Assume direct pushes to `main` are forbidden.
- Changes to `main` must go through a PR and use squash merge.
- PR threads must be resolved before merge.
- The required checks on `main` are:
  - `Quality Checks`
  - `Test Backend`
  - `Build Backend (linux, amd64)`
  - `Build Backend (linux, arm64)`
  - `Docker Build`
  - `Dependency Review`
  - `CodeQL / actions`
  - `CodeQL / go`
- The ruleset also enforces strict status checks, linear history, no force-pushes, and no branch deletion on `main`.

## Code Scanning Notes

- `.github/workflows/codeql.yml` is the repo-owned CodeQL source of truth.
- Keep CodeQL check names stable and unique. The required names are `CodeQL / actions` and `CodeQL / go`.
- Do not add ambiguous required checks such as `Analyze (go)`; GitHub can emit duplicate names from platform-managed workflows.
- If GitHub shows an extra neutral `CodeQL` check or a stray `Analyze (go)` run, treat that as platform noise unless it becomes a blocking required check.

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
