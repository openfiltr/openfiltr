# API Authentication Guide

OpenFiltr supports multiple authentication methods: JWT tokens, API tokens, and cookie-based sessions.

## Table of Contents

- [JWT Authentication](#jwt-authentication)
- [API Token Authentication](#api-token-authentication)
- [Cookie-Based Authentication](#cookie-based-authentication)
- [JWT vs API Tokens](#jwt-vs-api-tokens)
- [Token Revocation](#token-revocation)

---

## JWT Authentication

JSON Web Tokens (JWT) are best for stateless authentication in single-page applications and mobile apps.

### Login Flow

```bash
# Login with username and password
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-password"}'
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 86400,
  "user": {
    "id": "1",
    "username": "admin",
    "email": "admin@example.com"
  }
}
```

### Using JWT Tokens

Include the token in the `Authorization` header:

```bash
curl -X GET http://localhost:3000/api/dns/zones \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Token Expiration

JWT tokens expire after 24 hours by default. To refresh:

```bash
# Refresh token
curl -X POST http://localhost:3000/api/auth/refresh \
  -H "Authorization: Bearer <current-token>"
```

---

## API Token Authentication

API tokens are best for server-to-server communication and CI/CD pipelines.

### Creating API Tokens

1. **Via Dashboard:**
   - Go to Settings → API Tokens
   - Click "Generate New Token"
   - Copy the token (shown only once)

2. **Via API:**

```bash
# Create API token (requires authenticated session)
curl -X POST http://localhost:3000/api/auth/token \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "My API Token", "expires_in": 31536000}'
```

**Response:**

```json
{
  "token": "of_live_abc123xyz789...",
  "name": "My API Token",
  "expires_at": "2027-03-19T12:00:00Z"
}
```

### Using API Tokens

Include the token in the `Authorization` header:

```bash
curl -X GET http://localhost:3000/api/dns/zones \
  -H "Authorization: APIKEY of_live_abc123xyz789..."
```

### Token Format

API tokens use the prefix `of_live_` for production tokens and `of_test_` for test tokens.

---

## Cookie-Based Authentication

Cookie-based authentication is best for browser-based applications.

### Login Flow

```bash
# Login (returns session cookie)
curl -c cookies.txt -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your-password"}'
```

### Using Cookies

```bash
# Subsequent requests with cookie
curl -b cookies.txt http://localhost:3000/api/dns/zones
```

### Logout

```bash
# Clear session
curl -b cookies.txt -X POST http://localhost:3000/api/auth/logout
```

---

## JWT vs API Tokens

| Feature | JWT | API Token |
|---------|-----|-----------|
| **Best For** | SPAs, Mobile apps | Server-to-server, CI/CD |
| **Expiration** | Configurable (default 24h) | Long-lived or permanent |
| **Stateless** | Yes | No |
| **Revocable** | Difficult | Easy |
| **Size** | Larger | Smaller |

### When to Use JWT

- Single-page applications (React, Vue, Angular)
- Mobile apps
- When you need stateless authentication

### When to Use API Tokens

- Server-side scripts
- CI/CD pipelines
- Third-party integrations
- When you need easy revocation

---

## Token Revocation

### Revoking JWT Tokens

JWT tokens cannot be directly revoked, but you can:

1. **Add to blocklist:**

```bash
# Block a token (requires admin)
curl -X POST http://localhost:3000/api/auth/token/block \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"token": "<jwt-to-block>"}'
```

2. **Change JWT secret** (invalidates all tokens):

```bash
# Set new secret
export OPENFILTR_JWT_SECRET="new-secret-value"
```

### Revoking API Tokens

```bash
# List your tokens
curl -X GET http://localhost:3000/api/auth/tokens \
  -H "Authorization: Bearer <jwt-token>"

# Revoke a specific token
curl -X DELETE http://localhost:3000/api/auth/token/<token-id> \
  -H "Authorization: Bearer <jwt-token>"
```

---

## Security Best Practices

1. **Never expose tokens in URLs**
   - ✅ `Authorization: Bearer <token>`
   - ❌ `GET /api?token=<token>`

2. **Use HTTPS in production**
   - Tokens are encrypted in transit

3. **Set appropriate expiration times**
   - JWT: 1 hour for sensitive operations
   - API tokens: 1 year for services, shorter for temporary access

4. **Rotate tokens regularly**
   - Generate new tokens periodically
   - Revoke unused tokens

5. **Store tokens securely**
   - Use secure storage (keychain, vault)
   - Never commit to version control

---

## Error Responses

### 401 Unauthorized

```json
{
  "error": "invalid_token",
  "message": "Token has expired"
}
```

### 403 Forbidden

```json
{
  "error": "insufficient_permissions",
  "message": "You don't have permission to access this resource"
}
```

---

## Related Documentation

- [Configuration Reference](./configuration.md)
- [API Reference](./api.md)
- [Installation Guide](./installation.md)
