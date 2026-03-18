# OpenFiltr — AI Agent Instructions

> These instructions are for AI coding agents (GitHub Copilot, Claude, etc.) working on the OpenFiltr codebase. Read this before making any changes.

## Project overview

OpenFiltr is a **self-hosted DNS filtering platform** built in Go with a React + TypeScript + Tailwind CSS frontend. It blocks unwanted domains (ads, trackers, malware) across an entire network by acting as a local DNS resolver.

### Key facts

- **Language**: Go 1.23 (backend), TypeScript + React 18 (frontend)
- **Storage**: SQLite 3 with WAL mode; embedded migrations
- **DNS library**: `github.com/miekg/dns`
- **HTTP router**: `github.com/go-chi/chi/v5`
- **Authentication**: JWT (`github.com/golang-jwt/jwt/v5`) + bcrypt API tokens
- **Frontend build**: Vite
- **CSS**: Tailwind CSS utility classes only — no custom CSS files
- **Tests**: table-driven Go tests; `t.Run()` subtests

## Language

**Use British English everywhere** — comments, documentation, error messages, commit messages, PR descriptions, and issue bodies. Exceptions: Go identifiers and JSON keys use camelCase/snake_case as per their conventions.

## Repository layout

```
cmd/server/          — main entrypoint (main.go)
internal/
  api/               — HTTP handlers and router (chi)
  auth/              — JWT + bcrypt + API token logic
  config/            — YAML config loader
  dns/               — UDP DNS server + forwarding
  storage/           — SQLite open + migrations
pkg/sdk/             — Public Go SDK (Apache 2.0) — keep stable
web/                 — Vite + React + TypeScript UI
openapi/             — OpenAPI 3.1 YAML spec
deploy/docker/       — Dockerfile + docker-compose
scripts/             — install.sh + create-issues.sh
.github/             — Workflows, templates, project setup
```

## Go conventions

### Error handling

- Always wrap errors with context: `fmt.Errorf("opening database: %w", err)`
- Return errors up the call stack; only log at the top level
- Never use `panic` except in `init()` or `TestMain`

### Logging

- Use `log/slog` (stdlib) only — never `fmt.Println` in production paths
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

### Database

- All IDs are UUIDs: `github.com/google/uuid`
- Timestamps: `DATETIME DEFAULT CURRENT_TIMESTAMP` in SQLite
- Use `?` placeholders — never string interpolation in SQL
- Always `defer rows.Close()` after a query
- Wrap multi-step operations in a transaction

### Auth

- `auth.ExtractToken(r)` reads from `Authorization: Bearer` header or `openfiltr_token` cookie
- Middleware sets claims in context: `ctx.Value(claimsKey).(*auth.Claims)`
- Use `currentUser(r)` helper to get the current user in a handler

## React / TypeScript conventions

- Strict TypeScript: `"strict": true` in tsconfig
- Functional components only — no class components
- `@tanstack/react-query` for all data fetching; no manual `useState` + `useEffect` for API calls
- Tailwind utility classes only — no CSS modules, no `styled-components`
- `lucide-react` for all icons
- Form validation: inline, not with a library
- Error states: show an inline error message near the relevant input

### API client

All API calls go through `src/lib/api.ts`. Add new endpoints there, not inline.

```typescript
// Example
export const blockRules = {
  list: (params?: ListParams) => api.get<ListResponse<BlockRule>>('/api/v1/filtering/block-rules', { params }),
  create: (data: CreateBlockRuleInput) => api.post<BlockRule>('/api/v1/filtering/block-rules', data),
};
```

## Testing

### Go

- Table-driven tests with `t.Run()` subtests
- Use `t.Helper()` in assertion helpers
- Use `testing/iotest` and `net/http/httptest` for handler tests
- No external test frameworks — stdlib `testing` package only
- Test file naming: `*_test.go` in the same package

### TypeScript

- Vitest for unit tests
- React Testing Library for component tests
- No snapshot tests

## What to avoid

- ❌ Modal-heavy UI — prefer inline editing and expandable rows
- ❌ Hardcoded secrets, credentials, or test tokens in source code
- ❌ Direct SQL in HTTP handlers — use a storage layer function
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

### Add a new database table

1. Add a migration string to `internal/storage/migrations.go`
2. Add the model struct in the relevant package
3. Implement CRUD functions
4. Add CRUD API endpoints

### Add a new frontend page

1. Create `web/src/pages/MyPage.tsx`
2. Add the route in `web/src/App.tsx`
3. Add the sidebar link in `web/src/components/Sidebar.tsx`
4. Fetch data with `useQuery` from `@tanstack/react-query`

## Running locally

```bash
# Backend
make build && ./openfiltr

# Frontend (in another terminal)
cd web && npm run dev

# Both with live reload
make dev  # requires air: go install github.com/air-verse/air@latest
```

## AI-specific notes

- When generating SQL, always use `?` parameter placeholders
- When generating new handlers, follow the existing `respond`/`respondError` pattern
- When generating new list endpoints, return `{"items":[...], "total":N}` — not a raw array
- When adding features, check the OpenAPI spec first to see if the endpoint is already documented
- All new user-visible text should be in British English
- Do not introduce new Go dependencies without checking existing ones first
- Do not introduce new npm packages without checking `web/package.json` first
