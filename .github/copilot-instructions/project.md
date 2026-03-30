## Before you begin any task

Complete these steps **before making any code changes**:

1. **Sync with `main`** — fetch and rebase/merge the latest `origin/main` into your working branch.

2. **Read `CHANGELOG.md`** — understand what has recently changed and why. Pay particular attention to entries added after the most recent version tag; they describe in-progress or recently landed work that affects the current codebase state.

3. **Read all ADRs in `docs/adr/`** — these record significant architectural decisions (for example the storage decisions in ADR-0001 and ADR-0003, and the British English requirement). Do not reverse or work around a decision recorded in an ADR without first creating a new ADR that explicitly supersedes it.

4. **Run the test suite** to establish the passing baseline before you write a single line of code:
   ```bash
   go test ./...
   ```
   If tests already fail on `main`, note which ones so you do not report them as regressions you introduced.
5. **Follow the Codex workflow README when you need to branch, push, and open a PR**:
   ```bash
   sed -n '1,240p' docs/codex-agents/README.md
   ```

> Skipping any step risks working from stale state, undoing deliberate decisions, or generating misleading test failures.

---

# OpenFiltr - AI Agent Instructions

> These instructions are for AI coding agents (GitHub Copilot, Claude, etc.) working on the OpenFiltr codebase. Read this before making any changes.

## Project overview

OpenFiltr is a **self-hosted DNS filtering platform** built in Go. The repository is currently backend-first. Treat the API, DNS, auth, config, storage, Docker, and workflow files as the source of truth. Do not assume a working React UI or SDK exists just because older docs or issues mention them.

### Key facts

- **Language**: Go 1.24 backend
- **Storage**: bbolt by default, PostgreSQL optional
- **DNS library**: `github.com/miekg/dns`
- **HTTP router**: `github.com/go-chi/chi/v5`
- **Authentication**: JWT (`github.com/golang-jwt/jwt/v5`) + bcrypt API tokens
- **Tests**: table-driven Go tests; `t.Run()` subtests

## Language

**Use British English everywhere**: comments, documentation, error messages, commit messages, PR descriptions, and issue bodies. Exceptions: Go identifiers and JSON keys use camelCase and snake_case as per their conventions.

## Repository layout

```
cmd/server/          - main entrypoint (main.go)
internal/
  api/               - HTTP handlers and router (chi)
  auth/              - JWT + bcrypt + API token logic
  config/            - YAML config loader
  dns/               - UDP DNS server + forwarding
  storage/           - storage seam, bbolt store, PostgreSQL bootstrap + migrations
openapi/             - OpenAPI 3.1 YAML spec
deploy/docker/       - Dockerfile + docker-compose
scripts/             - install scripts and release helpers
.github/             - Workflows and templates
```

## Go conventions

### Error handling

- Always wrap errors with context: `fmt.Errorf("opening database: %w", err)`
- Return errors up the call stack; only log at the top level
- Never use `panic` except in `init()` or `TestMain`

### Logging

- Use `log/slog` (stdlib) only. Never `fmt.Println` in production paths
- Structured logging: `slog.Info("message", "key", value)`
- Never log passwords, JWT secrets, or raw API tokens

### HTTP handlers

Every handler follows this pattern:
```go
func (h *Handler) CreateBlockRule(w http.ResponseWriter, r *http.Request) {
    var req struct { Pattern string `json:"pattern"` }
    if err := decode(r, &req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    // ... logic ...
    respond(w, http.StatusCreated, result)
}
```

- Use `respond(w, status, data)` for all success responses
- Use `respondError(w, status, message)` for all error responses
- All list endpoints return `{"items": [...], "total": N}`
- All error responses return `{"error": {"message": "..."}}`

### Persistence

- All IDs are UUIDs: `github.com/google/uuid`
- bbolt is the default embedded backend
- PostgreSQL remains supported through `storage.database_url` or `OPENFILTR_DATABASE_URL`
- Local embedded storage uses `storage.database_path` or `OPENFILTR_DATABASE_PATH`
- New persistence work should consider both the bbolt and PostgreSQL paths unless the change is explicitly backend-specific
- Write SQL so it remains valid for the PostgreSQL compatibility path. Use the existing storage helpers where needed
- Always `defer rows.Close()` after a query
- Wrap multi-step operations in a transaction

### Auth

- `auth.ExtractToken(r)` reads from `Authorization: Bearer` header or `openfiltr_token` cookie
- Middleware sets claims in context: `ctx.Value(claimsKey).(*auth.Claims)`
- Use `currentUser(r)` helper to get the current user in a handler

## Testing

### Go

- Table-driven tests with `t.Run()` subtests
- Use `t.Helper()` in assertion helpers
- Use `testing/iotest` and `net/http/httptest` for handler tests
- No external test frameworks. Use the stdlib `testing` package only
- Test file naming: `*_test.go` in the same package

## What to avoid

- ❌ Hardcoded secrets, credentials, or test tokens in source code
- ❌ Direct SQL in HTTP handlers. Use a storage layer function
- ❌ Global mutable state in tests
- ❌ Disabling TLS verification in production paths
- ❌ `AllowedOrigins: ["*"]` with `AllowCredentials: true` in production
- ❌ Logging raw JWT secrets or password hashes

## Security checklist for every PR

- [ ] No new secrets introduced
- [ ] All user input validated before use in SQL or DNS queries
- [ ] All destructive actions logged to `audit_events`
- [ ] New endpoints added to the OpenAPI spec
- [ ] New endpoints protected by `AuthMiddleware` unless explicitly public

## Common tasks

### Add a new API endpoint

1. Add the route to `internal/api/router.go`
2. Add the handler method to the appropriate handler file
3. Add the OpenAPI path/schema to `openapi/openapi.yaml`
4. Write a table-driven test in `internal/api/*_test.go`

### Add new persistent data

1. Add the view or model struct in the relevant package
2. Update the bbolt store if the data belongs in the default backend
3. Add or update SQL migrations if PostgreSQL compatibility needs the same data
4. Implement CRUD functions without pushing raw SQL into handlers
5. Add CRUD API endpoints if the data is part of the public surface

## Running locally

```bash
# Backend
make build && ./openfiltr
```

## AI-specific notes

- When generating SQL, follow the repository's PostgreSQL compatibility conventions and existing storage helpers
- When generating new handlers, follow the existing `respond`/`respondError` pattern
- When generating new list endpoints, return `{"items":[...], "total":N}`, not a raw array
- When adding features, check the OpenAPI spec first to see if the endpoint is already documented
- All new user-visible text should be in British English
- Do not introduce new Go dependencies without checking existing ones first
