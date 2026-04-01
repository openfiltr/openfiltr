# OpenFiltr API Authentication Guide

This guide explains how to authenticate with the OpenFiltr API using both JWT tokens and API tokens.

## Table of Contents

- [Overview](#overview)
- [Initial Setup](#initial-setup)
- [JWT Authentication](#jwt-authentication)
- [API Token Authentication](#api-token-authentication)
- [Cookie-Based Authentication](#cookie-based-authentication)
- [When to Use JWT vs API Tokens](#when-to-use-jwt-vs-api-tokens)
- [Token Revocation](#token-revocation)
- [Security Best Practices](#security-best-practices)

## Overview

OpenFiltr supports two primary authentication methods:

1. **JWT (JSON Web Tokens)** - For browser-based sessions and temporary access
2. **API Tokens** - For programmatic access, automation, and long-term integrations

Both methods use the `Authorization` header with the `Bearer` scheme, or can be passed via HTTP cookies for browser sessions.

## Initial Setup

Before you can authenticate, you must create the initial admin user:

### Create Admin User

```bash
curl -X POST https://your-openfiltr-instance/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-secure-password"
  }'
```

**Response:**
```json
{
  "message": "setup complete"
}
```

**Note:** This endpoint is only available when no users exist in the system. Once a user is created, this endpoint returns a 409 Conflict error.

## JWT Authentication

JWT tokens are ideal for browser sessions and short-lived access. They contain encoded user information and have an expiration time configured by the server.

### Login to Get JWT Token

```bash
curl -X POST https://your-openfiltr-instance/api/v1/auth/login \
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

The response includes:
- `token` - The JWT token to use for authentication
- `username` - The authenticated username
- `role` - The user's role (typically "admin")

### Using JWT Token

**Option 1: Authorization Header (Recommended for API clients)**

```bash
curl -X GET https://your-openfiltr-instance/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Option 2: Cookie (For browser sessions)**

The login endpoint automatically sets an HTTP-only cookie named `openfiltr_token`. Browsers will include this cookie automatically in subsequent requests.

```bash
# Using cookie jar for session management
curl -X POST https://your-openfiltr-instance/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-secure-password"}' \
  -c cookies.txt

# Subsequent requests use the cookie
curl -X GET https://your-openfiltr-instance/api/v1/auth/me \
  -b cookies.txt
```

### Get Current User Info

```bash
curl -X GET https://your-openfiltr-instance/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**
```json
{
  "id": "uuid-here",
  "username": "admin",
  "role": "admin"
}
```

### Logout

```bash
curl -X POST https://your-openfiltr-instance/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

This clears the authentication cookie and invalidates the session on the client side.

## API Token Authentication

API tokens are designed for automation, scripts, and long-term integrations. They are more suitable for server-to-server communication and don't require re-authentication.

### Create an API Token

```bash
curl -X POST https://your-openfiltr-instance/api/v1/auth/tokens \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "automation-script",
    "expires_at": "2026-12-31T23:59:59Z"
  }'
```

**Parameters:**
- `name` (required) - A descriptive name for the token
- `expires_at` (optional) - Expiration time in RFC3339 format

**Response:**
```json
{
  "id": "uuid-here",
  "name": "automation-script",
  "token": "oft_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}
```

**Important:** The `token` value is only shown once. Store it securely as it cannot be retrieved again.

### List API Tokens

```bash
curl -X GET https://your-openfiltr-instance/api/v1/auth/tokens \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**
```json
{
  "items": [
    {
      "id": "uuid-here",
      "name": "automation-script",
      "scopes": "",
      "last_used_at": "2026-04-01T10:30:00Z",
      "expires_at": "2026-12-31T23:59:59Z",
      "created_at": "2026-04-01T10:00:00Z"
    }
  ],
  "total": 1
}
```

### Using API Token

```bash
curl -X GET https://your-openfiltr-instance/api/v1/system/status \
  -H "Authorization: Bearer oft_YOUR_API_TOKEN"
```

API tokens use the same `Authorization: Bearer` header as JWT tokens. The server automatically detects which type of token is being used.

### Example: Using API Token in a Script

```bash
#!/bin/bash
API_TOKEN="oft_1234567890abcdef..."

# Get system status
curl -X GET https://your-openfiltr-instance/api/v1/system/status \
  -H "Authorization: Bearer $API_TOKEN"

# Create a block rule
curl -X POST https://your-openfiltr-instance/api/v1/filtering/block-rules \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "ads.example.com",
    "enabled": true
  }'
```

## Cookie-Based Authentication

For browser-based sessions, OpenFiltr uses cookies with CSRF protection for enhanced security.

### How It Works

1. **Login**: The `/api/v1/auth/login` endpoint sets two cookies:
   - `openfiltr_token` - HTTP-only cookie containing the JWT token
   - CSRF token cookie

2. **CSRF Protection**: For state-changing operations (POST, PUT, DELETE), include the CSRF token in:
   - `X-CSRF-Token` header
   - Or `X-CSRF-Token` cookie

### Browser Example

```javascript
// Login
const response = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({username: 'admin', password: 'your-password'}),
  credentials: 'include' // Important: include cookies
});

const data = await response.json();
const csrfToken = response.headers.get('X-CSRF-Token');

// Subsequent requests with CSRF protection
const protectedResponse = await fetch('/api/v1/filtering/block-rules', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': csrfToken
  },
  credentials: 'include',
  body: JSON.stringify({pattern: 'ads.example.com', enabled: true})
});
```

### Curl Example with Cookies

```bash
# Login and save cookies
curl -X POST https://your-openfiltr-instance/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}' \
  -c cookies.txt \
  -D headers.txt

# Extract CSRF token from response headers
CSRF_TOKEN=$(grep -i 'X-CSRF-Token' headers.txt | cut -d' ' -f2 | tr -d '\r')

# Make protected request with CSRF token
curl -X POST https://your-openfiltr-instance/api/v1/filtering/block-rules \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -b cookies.txt \
  -d '{"pattern":"ads.example.com","enabled":true}'
```

## When to Use JWT vs API Tokens

### Use JWT Tokens When:
- **Browser-based sessions** - Users log in through a web interface
- **Short-lived access** - Token expiration is desirable
- **User context required** - You need user identity in each request
- **Interactive workflows** - Users actively interact with the application

**Advantages:**
- Automatic expiration for security
- User context embedded in token
- Compatible with browser cookie storage

**Limitations:**
- Requires periodic re-authentication
- Not ideal for long-running automation

### Use API Tokens When:
- **Automation and scripts** - CI/CD pipelines, scheduled jobs
- **Server-to-server communication** - Microservices or integrations
- **Long-term integrations** - Third-party tools that need persistent access
- **API clients** - Mobile apps, desktop applications, CLI tools

**Advantages:**
- No expiration (or custom expiration)
- Can be revoked without affecting user password
- Easier to manage for multiple integrations
- Token visible only once during creation (more secure)

**Limitations:**
- Must be stored securely
- No automatic expiration (unless configured)
- Requires manual rotation for security

### Decision Matrix

| Use Case | Recommended Method |
|----------|-------------------|
| Web UI login | JWT with cookie |
| Mobile app | API Token |
| CI/CD pipeline | API Token |
| Scheduled scripts | API Token |
| Third-party integration | API Token |
| Temporary access | JWT |
| Long-running service | API Token |

## Token Revocation

### Revoke API Tokens

You can revoke (delete) API tokens that are no longer needed:

```bash
curl -X DELETE https://your-openfiltr-instance/api/v1/auth/tokens/{token_id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**
```json
{
  "message": "deleted"
}
```

**Note:** You need the token `id` (not the token value) to delete it. Use the `GET /api/v1/auth/tokens` endpoint to list tokens and get their IDs.

### JWT Logout

While JWT tokens have expiration, you can explicitly logout:

```bash
curl -X POST https://your-openfiltr-instance/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

This clears the session cookie. Note that JWT tokens are stateless and cannot be truly revoked server-side until they expire.

### Security Implications

- **API Tokens**: Immediately revoked upon deletion
- **JWT Tokens**: Cannot be revoked server-side until expiration (configured by `AUTH_TOKEN_EXPIRY` environment variable)
- **Best Practice**: Use short expiration times for JWT and longer expiration (or no expiration) for API tokens with manual rotation

## Security Best Practices

### 1. Token Storage

- **JWT in browsers**: Use HTTP-only cookies to prevent XSS attacks
- **API Tokens**: Store in environment variables or secure secret management systems
- **Never**: Store tokens in code repositories, logs, or client-side localStorage

### 2. Token Rotation

```bash
# 1. Create new API token
NEW_TOKEN=$(curl -s -X POST https://your-openfiltr-instance/api/v1/auth/tokens \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"new-automation-token"}' | jq -r '.token')

# 2. Update your application to use the new token
export OPENFILTR_API_TOKEN="$NEW_TOKEN"

# 3. Revoke the old token
curl -X DELETE https://your-openfiltr-instance/api/v1/auth/tokens/$OLD_TOKEN_ID \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### 3. Scope Management

- Create separate API tokens for different purposes (e.g., monitoring, automation, reporting)
- Name tokens descriptively to track usage
- Set expiration dates for temporary integrations
- Regularly audit token usage via the list endpoint

### 4. Network Security

- Always use HTTPS in production
- Use secure cookie flags (`Secure`, `HttpOnly`, `SameSite=Strict`)
- Consider IP whitelisting for API token usage
- Monitor authentication logs for suspicious activity

### 5. Password Requirements

- Minimum 8 characters (enforced by the API)
- Use strong, unique passwords for admin accounts
- Consider implementing password rotation policies

### 6. Monitoring and Auditing

- Use the `/api/v1/auth/tokens` endpoint to audit active tokens
- Check the `last_used_at` timestamp to identify unused tokens
- Monitor the audit log endpoint for authentication events:
  ```bash
  curl -X GET https://your-openfiltr-instance/api/v1/audit \
    -H "Authorization: Bearer YOUR_JWT_TOKEN"
  ```

## Common Patterns

### CI/CD Pipeline Integration

```yaml
# GitHub Actions example
name: Update Block Rules
on:
  schedule:
    - cron: '0 0 * * *'

jobs:
  update-rules:
    runs-on: ubuntu-latest
    steps:
      - name: Fetch block list
        run: |
          curl -X POST https://openfiltr.example.com/api/v1/filtering/sources/{id}/refresh \
            -H "Authorization: Bearer ${{ secrets.OPENFILTR_API_TOKEN }}"
```

### Monitoring Script

```bash
#!/bin/bash
# Health check with API token
API_TOKEN="oft_your_token_here"

response=$(curl -s -w "%{http_code}" \
  -X GET https://openfiltr.example.com/api/v1/system/status \
  -H "Authorization: Bearer $API_TOKEN")

http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" -eq 200 ]; then
  echo "OpenFiltr is healthy"
  echo "$body" | jq .
else
  echo "OpenFiltr health check failed: $http_code"
  exit 1
fi
```

## Troubleshooting

### Invalid Credentials (401)

```bash
# Check your credentials
curl -X POST https://your-openfiltr-instance/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"wrong-password"}'

# Response: {"error":{"message":"invalid credentials"}}
```

**Solutions:**
- Verify username and password
- Ensure user exists (run setup if needed)
- Check for typos in credentials

### Token Expired

JWT tokens have an expiration time. If you receive 401 errors after some time:

```bash
# Re-login to get a new JWT
curl -X POST https://your-openfiltr-instance/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'
```

### CSRF Token Missing

For browser sessions making POST/PUT/DELETE requests:

```bash
# Ensure CSRF token is included in headers
curl -X POST https://your-openfiltr-instance/api/v1/filtering/block-rules \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_CSRF_TOKEN" \
  -b cookies.txt \
  -d '{"pattern":"ads.example.com","enabled":true}'
```

### API Token Not Found (404)

```bash
# Verify token exists in your token list
curl -X GET https://your-openfiltr-instance/api/v1/auth/tokens \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Additional Resources

- [OpenAPI Specification](/openapi.yaml) - Full API documentation
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute to OpenFiltr
- [Security Policy](../SECURITY.md) - Reporting security vulnerabilities