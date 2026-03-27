# API Authentication Guide

OpenFiltr supports three authentication methods: JWT tokens, long-lived API tokens, and browser-based session cookies.

## Setup (first run)

Before any authentication works, you must create the initial admin user:

```bash
curl -X POST http://localhost:3000/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-secure-password"}'
```

Returns `409 Conflict` if an admin user already exists.

## JWT Authentication

### Login

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-secure-password"}'
```

Response:
```json
{
  "token": "eyJhbGciOi...",
  "expires_at": "2026-03-28T12:00:00Z"
}
```

### Using the JWT

Include the token in the `Authorization` header for all subsequent requests:

```bash
curl http://localhost:3000/api/v1/users/me \
  -H "Authorization: Bearer eyJhbGciOi..."
```

**When to use JWT:** Short-lived sessions, programmatic access where tokens can be refreshed.

## API Token Authentication

### Create a token

```bash
curl -X POST http://localhost:3000/api/v1/auth/tokens \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{"name": "ci-pipeline", "expiry_days": 90}'
```

Response:
```json
{
  "token": "oft_abc123...",
  "name": "ci-pipeline",
  "created_at": "2026-03-27T12:00:00Z",
  "expires_at": "2026-06-25T12:00:00Z"
}
```

### Using API tokens

API tokens (`oft_` prefix) are used the same way as JWTs:

```bash
curl http://localhost:3000/api/v1/lists \
  -H "Authorization: Bearer oft_abc123..."
```

### Revoke a token

```bash
curl -X DELETE http://localhost:3000/api/v1/auth/tokens/<token-id> \
  -H "Authorization: Bearer oft_abc123..."
```

**When to use API tokens:** Long-lived access for CI/CD pipelines, external integrations, or services that need persistent access without frequent logins.

## Browser Session (Cookie-based)

### Login in browser

The login endpoint also sets cookies for browser-based usage:

- `openfiltr_token` — HttpOnly session cookie containing the JWT
- `openfiltr_csrf` — CSRF protection cookie

The response includes an `X-CSRF-Token` header that must be echoed on state-changing requests:

```javascript
// After login
const csrfToken = response.headers.get('X-CSRF-Token');

// On subsequent POST/PUT/DELETE requests
fetch('/api/v1/lists', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': csrfToken
  },
  body: JSON.stringify({ name: 'my-list' })
});
```

**When to use cookies:** Web UIs and browser-based clients where the session cookie is automatically sent.

## Choosing an authentication method

| Method | Lifetime | Best for |
|--------|----------|----------|
| JWT | Short-lived (24h default) | User sessions, interactive CLI |
| API Token | Long-lived (configurable) | CI/CD, automation, services |
| Cookie | Session-based | Web browsers, SPAs |
