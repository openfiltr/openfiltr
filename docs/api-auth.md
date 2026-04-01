# API Authentication Guide

This guide explains how to authenticate with the OpenFiltr API using both JWT tokens (for browser sessions) and API tokens (for programmatic access).

## Table of Contents

- [Authentication Methods](#authentication-methods)
- [JWT Token Authentication](#jwt-token-authentication)
- [API Token Authentication](#api-token-authentication)
- [Cookie-Based Authentication](#cookie-based-authentication)
- [When to Use JWT vs API Tokens](#when-to-use-jwt-vs-api-tokens)
- [Error Handling](#error-handling)

---

## Authentication Methods

OpenFiltr supports two primary authentication methods:

1. **JWT Tokens** - For browser-based sessions with automatic cookie management
2. **API Tokens** - For programmatic access with manual token management (prefix: `oft_`)

Both methods provide the same level of access to API endpoints, but differ in how they're obtained, used, and managed.

---

## JWT Token Authentication

JWT tokens are ideal for browser-based applications where users log in through a web interface. The authentication flow includes automatic cookie management and CSRF protection.

### Initial Setup

Before you can log in, you need to create the initial admin user:

```bash
# Check if setup is needed
curl -X GET https://your-openfiltr-instance/api/v1/system/health

# Create initial admin user
curl -X POST https://your-openfiltr-instance/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-secure-password-here"
  }'
```

**Response:**
```json
{
  "message": "setup complete"
}
```

**Note:** The setup endpoint is only available when no users exist in the system.

### Login Flow

Authenticate with your username and password to receive a JWT token:

```bash
curl -X POST https://your-openfiltr-instance/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "username": "admin",
    "password": "your-password-here"
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

The response includes:
- `token` - The JWT token for subsequent API calls
- `username` - The authenticated user's username
- `role` - The user's role (e.g., "admin")

**Cookie Behavior:**
- The server automatically sets an HTTP-only cookie named `openfiltr_token`
- Cookie attributes: `HttpOnly`, `Secure`, `SameSite=Strict`
- Cookie expiration matches the configured token expiry time (default: 24 hours)

### Using JWT Tokens

After login, browsers automatically include the cookie with subsequent requests. For curl or other HTTP clients, use the saved cookies:

```bash
# Using saved cookies
curl -X GET https://your-openfiltr-instance/api/v1/auth/me \
  -b cookies.txt

# Or manually using the Authorization header
curl -X GET https://your-openfiltr-instance/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "admin",
  "role": "admin"
}
```

### Logout

To log out and invalidate the session cookie:

```bash
curl -X POST https://your-openfiltr-instance/api/v1/auth/logout \
  -b cookies.txt
```

**Response:**
```json
{
  "message": "logged out"
}
```

This clears the authentication cookie and CSRF cookie.

---

## API Token Authentication

API tokens are designed for programmatic access, automation, CI/CD pipelines, and integrations. They use the `oft_` prefix for easy identification.

### Creating API Tokens

Create a new API token through the authenticated API:

```bash
# Create a token with no expiration
curl -X POST https://your-openfiltr-instance/api/v1/auth/tokens \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "CI/CD Pipeline"
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "name": "CI/CD Pipeline",
  "token": "oft_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"
}
```

**Important:** Save the `token` value securely. You won't be able to see it again.

#### Token with Expiration

Create a token that expires on a specific date:

```bash
curl -X POST https://your-openfiltr-instance/api/v1/auth/tokens \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Temporary Access",
    "expires_at": "2026-12-31T23:59:59Z"
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "name": "Temporary Access",
  "token": "oft_b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8"
}
```

The `expires_at` field uses RFC3339 format.

### Listing API Tokens

View all your API tokens:

```bash
curl -X GET https://your-openfiltr-instance/api/v1/auth/tokens \
  -H "Authorization: Bearer your-jwt-token"
```

**Response:**
```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "CI/CD Pipeline",
      "scopes": "",
      "last_used_at": "2026-03-15T10:30:00Z",
      "expires_at": null,
      "created_at": "2026-03-01T12:00:00Z"
    }
  ],
  "total": 1
}
```

**Note:** The actual token value is never returned in the list for security reasons.

### Using API Tokens

Use the token in the `Authorization` header with the `Bearer` scheme:

```bash
curl -X GET https://your-openfiltr-instance/api/v1/system/status \
  -H "Authorization: Bearer oft_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"
```

API tokens bypass CSRF protection since they're designed for automated access.

### Token Revocation

Delete (revoke) an API token when it's no longer needed:

```bash
curl -X DELETE https://your-openfiltr-instance/api/v1/auth/tokens/550e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer your-jwt-token"
```

**Response:**
```json
{
  "message": "deleted"
}
```

After revocation, the token immediately becomes invalid for all future requests.

---

## Cookie-Based Authentication

When using JWT tokens through a browser, OpenFiltr automatically manages cookies and provides CSRF protection.

### How It Works

1. **Login**: Server sets two cookies:
   - `openfiltr_token` - HTTP-only cookie containing the JWT
   - `openfiltr_csrf` - Cookie containing the CSRF token

2. **Subsequent Requests**: Browser automatically sends the auth cookie

3. **State-Changing Requests**: For POST, PUT, DELETE, and PATCH requests, you must include the CSRF token:
   ```html
   <script>
   // Get CSRF token from cookie or initial login response
   const csrfToken = getCookie('openfiltr_csrf');
   
   fetch('/api/v1/filtering/block-rules', {
     method: 'POST',
     headers: {
       'Content-Type': 'application/json',
       'X-CSRF-Token': csrfToken
     },
     credentials: 'include', // Include cookies
     body: JSON.stringify({
       domain: 'example.com',
       enabled: true
     })
   });
   </script>
   ```

### CSRF Token Handling

The CSRF token is returned:
- In the `X-CSRF-Token` response header during login
- In the `openfiltr_csrf` cookie

**Example with curl:**
```bash
# Login and save cookies
curl -X POST https://your-openfiltr-instance/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"username": "admin", "password": "your-password"}' \
  -D headers.txt

# Extract CSRF token from headers or cookies
# Then include it in state-changing requests
curl -X POST https://your-openfiltr-instance/api/v1/filtering/block-rules \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: extracted-csrf-token" \
  -b cookies.txt \
  -d '{"domain": "malicious.com", "enabled": true}'
```

### When CSRF Protection Applies

CSRF protection is enforced when:
- Using cookie-based authentication (not Bearer tokens)
- Making state-changing requests (POST, PUT, DELETE, PATCH)

CSRF protection is **not** required when:
- Using API tokens with the `Authorization: Bearer` header
- Making read-only requests (GET, HEAD, OPTIONS)

---

## When to Use JWT vs API Tokens

### Use JWT Tokens When:

- **Browser-based applications** - Web UIs that users interact with directly
- **Session-based workflows** - Temporary access that expires automatically
- **User-initiated actions** - Actions performed by logged-in users
- **Development/testing** - Quick authentication during development

**Advantages:**
- Automatic cookie management in browsers
- Built-in expiration (configurable, default 24 hours)
- CSRF protection for browser security

**Limitations:**
- Requires cookie support
- Session management tied to browser
- CSRF token handling for state changes

### Use API Tokens When:

- **Automated scripts** - Cron jobs, scheduled tasks, automation
- **CI/CD pipelines** - GitHub Actions, Jenkins, GitLab CI, etc.
- **Integrations** - Third-party services, monitoring tools
- **Service-to-service** - Backend services communicating with OpenFiltr
- **Long-lived access** - Persistent authentication without expiration
- **Command-line tools** - Custom CLI applications

**Advantages:**
- No cookie or CSRF handling required
- Can be stored securely in environment variables
- Revocable without affecting user sessions
- Can set custom expiration dates
- Easily identifiable by `oft_` prefix

**Limitations:**
- Token value only shown once at creation
- Must manage token storage securely
- No automatic expiration (unless specified)

### Comparison Table

| Feature | JWT Tokens | API Tokens |
|---------|-----------|-----------|
| **Authentication Method** | Cookie or Bearer header | Bearer header only |
| **CSRF Protection** | Required for state changes | Not required |
| **Expiration** | Automatic (default 24h) | Optional |
| **Revocation** | Logout clears cookie | Delete by ID |
| **Token Visibility** | Can be retrieved | Shown once only |
| **Use Case** | Browser sessions | Automation/API access |
| **Prefix** | None | `oft_` |

---

## Error Handling

### Common Authentication Errors

#### 401 Unauthorized - Missing or Invalid Token

```json
{
  "error": {
    "message": "authentication required"
  }
}
```

**Cause:** No authentication token provided

**Solution:** Include `Authorization: Bearer <token>` header or ensure cookies are sent

---

#### 401 Unauthorized - Invalid or Expired Token

```json
{
  "error": {
    "message": "invalid or expired token"
  }
}
```

**Cause:** Token is invalid, expired, or has been revoked

**Solution:** 
- For JWT: Log in again to get a new token
- For API tokens: Check if token was revoked, create a new one

---

#### 403 Forbidden - CSRF Token Required

```json
{
  "error": {
    "message": "CSRF token required"
  }
}
```

**Cause:** Using cookie-based authentication without CSRF token for state-changing request

**Solution:** 
- Include `X-CSRF-Token` header with CSRF token value
- Or use API token with `Authorization: Bearer` header

---

#### 400 Bad Request - Invalid expires_at Format

```json
{
  "error": {
    "message": "invalid expires_at (use RFC3339)"
  }
}
```

**Cause:** API token expiration date not in RFC3339 format

**Solution:** Use format like `2026-12-31T23:59:59Z`

**Example:**
```bash
# Correct format
curl -X POST https://your-openfiltr-instance/api/v1/auth/tokens \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "expires_at": "2026-12-31T23:59:59Z"}'
```

---

#### 409 Conflict - Setup Already Completed

```json
{
  "error": {
    "message": "setup already completed"
  }
}
```

**Cause:** Attempting to use setup endpoint when users already exist

**Solution:** Use the standard login endpoint instead

---

## Security Best Practices

1. **Store API tokens securely** - Use environment variables, secret managers, or encrypted storage
2. **Use HTTPS** - Always transmit tokens over encrypted connections
3. **Set expiration dates** - For API tokens with limited-time access needs
4. **Revoke unused tokens** - Regularly audit and delete old API tokens
5. **Monitor token usage** - Check `last_used_at` to identify unused tokens
6. **Protect CSRF tokens** - Treat CSRF tokens with same care as passwords
7. **Use strong passwords** - Minimum 8 characters for user accounts

---

## Quick Reference

### JWT Authentication Flow

```bash
# 1. Initial setup (first time only)
curl -X POST https://api.example.com/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "secure-password"}'

# 2. Login
curl -X POST https://api.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"username": "admin", "password": "secure-password"}'

# 3. Use API (with cookies)
curl -X GET https://api.example.com/api/v1/system/status \
  -b cookies.txt

# 4. Logout
curl -X POST https://api.example.com/api/v1/auth/logout \
  -b cookies.txt
```

### API Token Authentication Flow

```bash
# 1. Create token (requires JWT auth first)
curl -X POST https://api.example.com/api/v1/auth/tokens \
  -H "Authorization: Bearer jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Automation"}'

# Save the returned token: oft_...

# 2. Use API with token
curl -X GET https://api.example.com/api/v1/system/status \
  -H "Authorization: Bearer oft_your-token-here"

# 3. List tokens
curl -X GET https://api.example.com/api/v1/auth/tokens \
  -H "Authorization: Bearer jwt-token"

# 4. Revoke token
curl -X DELETE https://api.example.com/api/v1/auth/tokens/{token-id} \
  -H "Authorization: Bearer jwt-token"
```