# PostgreSQL Storage And Agent Guidance Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace SQLite with PostgreSQL as the only supported database backend and add a root `AGENTS.md` file that gives coding agents accurate repository and workflow guidance.

**Architecture:** Keep the current `database/sql` shape, but make it explicitly PostgreSQL-only. Move schema creation into versioned SQL migration files, convert runtime queries to PostgreSQL syntax, and update config, Docker, CI, and docs to match. Add one root `AGENTS.md` that documents the real repo layout, coding expectations, and worktree usage.

**Tech Stack:** Go, PostgreSQL, `database/sql`, `github.com/lib/pq` or `github.com/jackc/pgx` stdlib driver, YAML, GitHub Actions, Docker Compose

---

## Chunk 1: Repo Guidance And Config Surface

### Task 1: Add Root Agent Guidance

**Files:**
- Create: `AGENTS.md`
- Modify: `README.md`
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Write the guidance document**

Include:
- repo truth and current limitations
- British English requirement
- main entry points and important directories
- worktree rule for local `.worktrees/`
- preferred verification commands
- warning that README may describe not-yet-present UI/SDK pieces

- [ ] **Step 2: Align human-facing docs**

Update any repo docs that contradict the new agent guidance in a way that would mislead contributors about storage or missing code.

- [ ] **Step 3: Verify the document is discoverable and coherent**

Run: `sed -n '1,240p' AGENTS.md`
Expected: the file explains repo structure, coding expectations, verification, and worktree usage clearly.

- [ ] **Step 4: Commit**

Run:
```bash
git add AGENTS.md README.md CONTRIBUTING.md
git commit -s -m "docs(repo): add agent guidance"
```

### Task 2: Replace SQLite Config With PostgreSQL Config

**Files:**
- Modify: `internal/config/config.go`
- Modify: `config/app.yaml.example`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Write the failing config expectation**

Add or extend config tests if any exist. If no test harness exists, add a focused Go test file for config loading that proves:
- default config exposes PostgreSQL settings
- `OPENFILTR_DATABASE_URL` overrides YAML config

- [ ] **Step 2: Run the focused config test to verify failure**

Run: `go test ./internal/config -run Test`
Expected: FAIL because config still exposes SQLite path settings.

- [ ] **Step 3: Implement PostgreSQL-first config**

Replace `storage.path` with a PostgreSQL DSN field or equivalent structure, update defaults, and wire the environment override to `OPENFILTR_DATABASE_URL`.

- [ ] **Step 4: Run the focused config test to verify pass**

Run: `go test ./internal/config -run Test`
Expected: PASS.

- [ ] **Step 5: Commit**

Run:
```bash
git add internal/config/config.go config/app.yaml.example cmd/server/main.go internal/config/*_test.go
git commit -s -m "feat(config): switch storage config to postgres"
```

## Chunk 2: Storage Bootstrap And Schema

### Task 3: Replace SQLite Driver And Embedded Migrations

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `internal/storage/db.go`
- Modify: `internal/storage/migrations.go`
- Create: `internal/storage/migrations/001_initial_schema.sql`

- [ ] **Step 1: Write the failing storage test**

Add a focused storage test that asserts:
- the migration runner can locate SQL migration files
- opening the database uses the PostgreSQL driver and rejects empty DSN input clearly

- [ ] **Step 2: Run the focused storage test to verify failure**

Run: `go test ./internal/storage -run Test`
Expected: FAIL because storage still uses `sqlite3` and embedded migrations.

- [ ] **Step 3: Implement PostgreSQL storage bootstrap**

Switch to a PostgreSQL driver, remove SQLite-specific connection flags and single-connection assumptions, and load ordered SQL migrations from disk.

- [ ] **Step 4: Move the schema into SQL files**

Create `001_initial_schema.sql` with PostgreSQL-native table definitions and indexes matching current application needs closely enough to avoid unnecessary handler churn.

- [ ] **Step 5: Run the focused storage test to verify pass**

Run: `go test ./internal/storage -run Test`
Expected: PASS.

- [ ] **Step 6: Commit**

Run:
```bash
git add go.mod go.sum internal/storage/db.go internal/storage/migrations.go internal/storage/migrations/*.sql internal/storage/*_test.go
git commit -s -m "feat(storage): add postgres migrations"
```

## Chunk 3: Query Port

### Task 4: Convert Application SQL To PostgreSQL Syntax

**Files:**
- Modify: `internal/auth/auth.go`
- Modify: `internal/dns/server.go`
- Modify: `internal/api/activity_handler.go`
- Modify: `internal/api/audit_handler.go`
- Modify: `internal/api/auth_handler.go`
- Modify: `internal/api/clients_handler.go`
- Modify: `internal/api/config_handler.go`
- Modify: `internal/api/dns_handler.go`
- Modify: `internal/api/filtering_handler.go`
- Modify: `internal/api/handler.go`

- [ ] **Step 1: Write failing targeted tests for representative handlers**

Cover at least:
- one auth path
- one CRUD handler
- one config import/export path
- one DNS log/write path

- [ ] **Step 2: Run the targeted tests to verify failure**

Run: `go test ./internal/auth ./internal/api ./internal/dns`
Expected: FAIL because of SQLite placeholders/functions and PostgreSQL incompatibilities.

- [ ] **Step 3: Port all SQL syntax**

Convert:
- `?` placeholders to positional PostgreSQL placeholders
- `datetime('now')` to `NOW()`
- wildcard and concatenation expressions to PostgreSQL-safe syntax
- upserts to PostgreSQL form

Keep handler behaviour otherwise unchanged.

- [ ] **Step 4: Run the targeted tests to verify pass**

Run: `go test ./internal/auth ./internal/api ./internal/dns`
Expected: PASS.

- [ ] **Step 5: Grep for SQLite leftovers**

Run:
```bash
rg -n "sqlite|sqlite3|datetime\\('now'\\)|\\?" internal
```
Expected: no remaining SQLite driver references or SQL placeholder leftovers that belong to SQL strings.

- [ ] **Step 6: Commit**

Run:
```bash
git add internal/auth/auth.go internal/dns/server.go internal/api/*.go
git commit -s -m "refactor(storage): port queries to postgres"
```

## Chunk 4: Tooling, Docker, And CI

### Task 5: Replace SQLite References In Runtime Tooling

**Files:**
- Modify: `deploy/docker/Dockerfile`
- Modify: `deploy/docker/docker-compose.yml`
- Modify: `examples/docker-compose.yml`
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: Write the failing environment expectation**

Add lightweight verification or smoke checks where practical, or failing grep assertions in CI-oriented shell tests if no better harness exists.

- [ ] **Step 2: Run the relevant verification to confirm failure**

Run:
```bash
rg -n "sqlite|libsqlite3|OPENFILTR_DB_PATH|openfiltr\\.db" deploy .github config README.md
```
Expected: matches still exist.

- [ ] **Step 3: Replace runtime references**

Update Docker, Compose, and workflows so PostgreSQL is the expected database and SQLite build dependencies are removed unless still genuinely required elsewhere.

- [ ] **Step 4: Re-run the verification**

Run:
```bash
rg -n "sqlite|libsqlite3|OPENFILTR_DB_PATH|openfiltr\\.db" deploy .github config README.md
```
Expected: no stale SQLite runtime references remain.

- [ ] **Step 5: Commit**

Run:
```bash
git add deploy/docker/Dockerfile deploy/docker/docker-compose.yml examples/docker-compose.yml .github/workflows/ci.yml .github/workflows/release.yml
git commit -s -m "ci(deploy): replace sqlite runtime references"
```

## Chunk 5: Docs, Final Verification, And PR

### Task 6: Update Product Docs And Finalise The Branch

**Files:**
- Modify: `README.md`
- Modify: `ROADMAP.md`
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Update docs to describe PostgreSQL truthfully**

Remove SQLite claims, document PostgreSQL setup expectations, and keep roadmap language consistent with the new direction.

- [ ] **Step 2: Run final repository checks**

Run:
```bash
git diff --check
ruby -e 'require "yaml"; Dir[".github/workflows/*.yml", "config/*.yaml*"].each { |f| YAML.load_file(f) rescue abort("bad yaml: #{f}") }; puts "YAML OK"'
rg -n "sqlite|sqlite3|OPENFILTR_DB_PATH|openfiltr\\.db" .
```
Expected:
- no diff formatting errors
- YAML parses
- no stale SQLite references remain except in historical/spec text if intentionally kept

- [ ] **Step 3: If Go is available, run full Go verification**

Run:
```bash
go test ./...
```
Expected: PASS.

- [ ] **Step 4: Commit**

Run:
```bash
git add README.md ROADMAP.md CONTRIBUTING.md
git commit -s -m "docs(storage): document postgres transition"
```

- [ ] **Step 5: Push and open PR**

Run:
```bash
git push -u origin codex/postgres-agent-guidance
gh pr create --base main --head codex/postgres-agent-guidance --title "Replace SQLite with PostgreSQL and add agent guidance" --body-file <prepared-pr-body>
```

