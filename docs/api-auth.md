# API Authentication Guide

This guide explains how to authenticate with the OpenFiltr API using both JWT tokens and long-lived API tokens.

## Overview

OpenFiltr supports two primary authentication methods:

| Method | Use Case | Lifetime |
|--------|----------|----------|
| **JWT Token** | Browser sessions, short-lived automation | Hours (configurable) |
| **API Token** | Scripts, CI/CD, integrations | Days to years (configurable) |

All authenticated endpoints require a bearer token in the `Authorization` header:

```http
Authorization: Bearer <token>
```

---

## Quick Start

### First-Time Setup

Before you can authenticate, you need to create the initial admin user:

```bash
curl -X POST http://localhost:3000/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-secure-password"
  }'
```

**Response:** `201 Created`

> **Note:** The setup endpoint returns `409 Conflict` if a user already exists.

### Login to Get a JWT

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-secure-password"
  }'
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "admin",
  "role": "admin"
}
```

Use the token in subsequent requests:

```bash
curl http://localhost:3000/api/v1/filtering/block-rules \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## Authentication Methods

### 1. JWT Token (Browser Sessions)

JWT tokens are designed for browser-based sessions and short-lived automation tasks.

#### How It Works

1. **Login:** Send credentials to `/api/v1/auth/login`
2. **Receive Token:** Get a JWT token in the response body
3. **Use Token:** Include it in the `Authorization: Bearer` header

JWT tokens have a configurable expiry time (default: 24 hours).

#### Login Example

```bash
# Login and extract the token
TOKEN=$(curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-password"}' \
  | jq -r '.token')

# Use the token for subsequent requests
curl http://localhost:3000/api/v1/filtering/block-rules \
  -H "Authorization: Bearer $TOKEN"
```

#### Logout

```bash
curl -X POST http://localhost:3000/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

#### When to Use JWT Tokens

- **Browser applications** with user login flows
- **Short-lived scripts** that run on-demand
- **Interactive sessions** where you want automatic expiration

---

### 2. Cookie-Based Authentication (Browser Sessions)

For browser applications, OpenFiltr also supports cookie-based authentication with CSRF protection.

#### How It Works

1. **Login:** The server sets an `openfiltr_token` HttpOnly cookie
2. **CSRF Protection:** The response includes an `X-CSRF-Token` header and `openfiltr_csrf` cookie
3. **State-Changing Requests:** Include the CSRF token in the `X-CSRF-Token` header

#### Example (Browser)

```javascript
// Login
const response = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username: 'admin', password: 'your-password' })
});

// Extract CSRF token from response header
const csrfToken = response.headers.get('X-CSRF-Token');

// Use CSRF token for state-changing requests
await fetch('/api/v1/filtering/block-rules', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': csrfToken
  },
  body: JSON.stringify({ pattern: 'ads.example.com', rule_type: 'wildcard' })
});
```

> **Important:** The CSRF token is required for all state-changing requests (POST, PUT, PATCH, DELETE) when using cookie-based authentication.

---

### 3. API Token (Long-Lived)

API tokens are designed for automation, scripts, CI/CD pipelines, and integrations.

#### Key Features

- Prefix: `oft_` (e.g., `oft_a1b2c3d4e5f6...`)
- Configurable expiry (days to years, or no expiry)
- Stored securely — shown only once at creation
- Can be revoked at any time

#### Create an API Token

```bash
# First, get a JWT token via login
JWT_TOKEN=$(curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-password"}' \
  | jq -r '.token')

# Create an API token
curl -X POST http://localhost:3000/api/v1/auth/tokens \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CI Deployment Token",
    "expires_at": "2027-01-01T00:00:00Z"
  }'
```

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "CI Deployment Token",
  "token": "oft_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "created_at": "2026-04-01T12:00:00Z",
  "expires_at": "2027-01-01T00:00:00Z"
}
```

> **Warning:** The `token` field is returned **only once**. Store it securely immediately.

#### Use an API Token

```bash
# Use the API token directly
API_TOKEN="oft_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"

curl http://localhost:3000/api/v1/filtering/block-rules \
  -H "Authorization: Bearer $API_TOKEN"
```

#### List Your API Tokens

```bash
curl http://localhost:3000/api/v1/auth/tokens \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Response:**

```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "CI Deployment Token",
      "scopes": "",
      "last_used_at": "2026-04-01T10:30:00Z",
      "expires_at": "2027-01-01T00:00:00Z",
      "created_at": "2026-04-01T12:00:00Z"
    }
  ],
  "total": 1
}
```

> **Note:** Raw token values are never returned after creation. You only see metadata.

#### Revoke an API Token

```bash
curl -X DELETE http://localhost:3000/api/v1/auth/tokens/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $JWT_TOKEN"
```

**Response:** `200 OK`

#### When to Use API Tokens

- **CI/CD pipelines** for automated deployments
- **Monitoring scripts** that run periodically
- **Integrations** with other tools or services
- **Long-running services** that need persistent authentication
- **Service-to-service** communication

---

## JWT vs API Tokens: When to Use Which?

| Scenario | Recommended Method |
|----------|-------------------|
| User login in a web app | JWT with cookie + CSRF |
| Short automation script | JWT token |
| CI/CD pipeline | API token |
| Scheduled monitoring job | API token |
| Interactive API exploration | JWT token |
| Production integration | API token |
| Mobile/desktop app | API token |

### JWT Tokens

✅ **Advantages:**
- Auto-expire for security
- Can be obtained via login flow
- Works well with browser sessions

❌ **Disadvantages:**
- Need to refresh periodically
- Tied to user session

### API Tokens

✅ **Advantages:**
- Long-lived (days, months, years)
- Can be scoped and named
- Easy to rotate or revoke
- No session dependency

❌ **Disadvantages:**
- Must be stored securely
- Only shown once at creation
- No auto-expiry (unless configured)

---

## Security Best Practices

### 1. Store Tokens Securely

```bash
# Good: Use environment variables
export OPENFILTR_API_TOKEN="oft_..."
curl -H "Authorization: Bearer $OPENFILTR_API_TOKEN" ...

# Bad: Hardcode in scripts
curl -H "Authorization: Bearer oft_a1b2c3..." ...
```

### 2. Use HTTPS in Production

```bash
# Production
curl https://api.yourdomain.com/api/v1/auth/login ...

# Development only
curl http://localhost:3000/api/v1/auth/login ...
```

### 3. Set Token Expiry

```bash
# Set a reasonable expiry for API tokens (e.g., 1 year)
curl -X POST http://localhost:3000/api/v1/auth/tokens \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production API Token",
    "expires_at": "2027-04-01T00:00:00Z"
  }'
```

### 4. Rotate Tokens Regularly

1. Create a new API token
2. Update your scripts/services with the new token
3. Revoke the old token

### 5. Revoke Compromised Tokens Immediately

```bash
# List tokens to find the one to revoke
curl http://localhost:3000/api/v1/auth/tokens \
  -H "Authorization: Bearer $JWT_TOKEN"

# Revoke immediately
curl -X DELETE http://localhost:3000/api/v1/auth/tokens/<token-id> \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### 6. Use Strong Passwords

For the initial admin user and any additional users, use strong, unique passwords:

```bash
# Generate a secure password
openssl rand -base64 24
```

---

## Example Workflows

### CI/CD Pipeline Example

```yaml
# .github/workflows/deploy.yml
name: Deploy Block Rules

on:
  push:
    paths:
      - 'block-rules.yaml'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy rules to OpenFiltr
        env:
          OPENFILTR_API_URL: https://dns.yourdomain.com
          OPENFILTR_API_TOKEN: ${{ secrets.OPENFILTR_API_TOKEN }}
        run: |
          # Read rules from file
          RULES=$(cat block-rules.yaml)

          # Apply rules via API
          curl -X POST "$OPENFILTR_API_URL/api/v1/filtering/block-rules" \
            -H "Authorization: Bearer $OPENFILTR_API_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$RULES"
```

### Monitoring Script Example

```bash
#!/bin/bash
# monitor-dns.sh

API_URL="https://dns.yourdomain.com"
API_TOKEN="${OPENFILTR_API_TOKEN:?Set OPENFILTR_API_TOKEN environment variable}"

# Get current block rules count
COUNT=$(curl -s "$API_URL/api/v1/filtering/block-rules" \
  -H "Authorization: Bearer $API_TOKEN" \
  | jq '.total')

echo "Current block rules: $COUNT"

# Check system health
HEALTH=$(curl -s "$API_URL/api/v1/system/health")
echo "System health: $HEALTH"
```

---

## Troubleshooting

### Invalid Credentials

```json
{
  "error": {
    "message": "invalid credentials"
  }
}
```

**Solution:** Check your username and password. Ensure the user exists.

### Token Expired

```json
{
  "error": {
    "message": "token is expired"
  }
}
```

**Solution:** Login again to get a new JWT token, or use an API token.

### Invalid Token

```json
{
  "error": {
    "message": "invalid token"
  }
}
```

**Solution:** Ensure the token is correctly formatted in the `Authorization: Bearer <token>` header.

### CSRF Token Required

```json
{
  "error": {
    "message": "CSRF token required"
  }
}
```

**Solution:** When using cookie-based authentication, include the `X-CSRF-Token` header with the value from the login response.

---

## API Reference Summary

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/auth/setup` | POST | Create initial admin user |
| `/api/v1/auth/login` | POST | Login to get JWT token |
| `/api/v1/auth/logout` | POST | Logout and clear session |
| `/api/v1/auth/me` | GET | Get current user info |
| `/api/v1/auth/tokens` | GET | List API tokens |
| `/api/v1/auth/tokens` | POST | Create new API token |
| `/api/v1/auth/tokens/{id}` | DELETE | Revoke an API token |

---

## Next Steps

- [OpenAPI Specification](../openapi/openapi.yaml) — Full API specification
- [README](../README.md) — Project overview and installation guide