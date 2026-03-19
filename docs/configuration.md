# Configuration Reference

Complete reference for all OpenFiltR configuration options.

## File Location

Default configuration file: `config/app.yaml`

You can also use environment variables to override YAML settings.

## Configuration Structure

```yaml
version: 1

server:
  listen_http: ":3000"
  listen_dns: ":53"
  base_url: "http://localhost:3000"

dns:
  cache_ttl: 300
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"

storage:
  database_url: "postgres://openfiltr:openfiltr@localhost:5432/openfiltr?sslmode=disable"

auth:
  token_expiry_hours: 24
```

## Server Options

### `server.listen_http`

| Property | Value |
|----------|-------|
| Type | string |
| Default | `:3000` |
| Env Variable | `OPENFILTR_HTTP_PORT` |

HTTP server listen address. Format: `host:port` or just `:port` for all interfaces.

**Examples:**
```yaml
server:
  listen_http: ":3000"       # All interfaces, port 3000
  listen_http: "127.0.0.1:3000"  # Localhost only
```

---

### `server.listen_dns`

| Property | Value |
|----------|-------|
| Type | string |
| Default | `:53` |
| Env Variable | `OPENFILTR_DNS_PORT` |

DNS server listen address. Requires root privileges on Linux/macOS.

**Examples:**
```yaml
server:
  listen_dns: ":53"           # All interfaces, port 53
  listen_dns: "0.0.0.0:53"   # Explicit interface
```

**Note:** On Linux, you may need to:
```bash
sudo setcap cap_net_bind_service=+ep /usr/local/bin/openfiltr
```

---

### `server.base_url`

| Property | Value |
|----------|-------|
| Type | string |
| Default | `http://localhost:3000` |
| Env Variable | `OPENFILTR_BASE_URL` |

Public-facing base URL for generating links and redirects.

**Examples:**
```yaml
server:
  base_url: "https://dns.example.com"
  base_url: "http://192.168.1.100:3000"
```

---

## DNS Options

### `dns.cache_ttl`

| Property | Value |
|----------|-------|
| Type | integer |
| Default | `300` |
| Env Variable | `OPENFILTR_DNS_CACHE_TTL` |

DNS response cache time-to-live in seconds.

**Examples:**
```yaml
dns:
  cache_ttl: 300    # 5 minutes
  cache_ttl: 60    # 1 minute
  cache_ttl: 3600  # 1 hour
```

---

### `dns.upstream_servers`

| Property | Value |
|----------|-------|
| Type | array of objects |
| Default | Cloudflare (1.1.1.1), Quad9 (9.9.9.9) |
| Env Variable | `OPENFILTR_UPSTREAM_DNS` |

Upstream DNS servers to forward queries to.

**Object Properties:**
- `name`: string - Display name for the server
- `address`: string - IP:port of the DNS server

**Examples:**
```yaml
dns:
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
    - name: Google
      address: "8.8.8.8:53"
```

---

## Storage Options

### `storage.database_url`

| Property | Value |
|----------|-------|
| Type | string |
| Default | `postgres://openfiltr:openfiltr@localhost:5432/openfiltr?sslmode=disable` |
| Env Variable | `DATABASE_URL` |

PostgreSQL database connection string.

**Format:**
```
postgres://username:password@host:port/database?sslmode=mode
```

**Examples:**
```yaml
storage:
  # Local development
  database_url: "postgres://openfiltr:openfiltr@localhost:5432/openfiltr?sslmode=disable"
  
  # With password
  database_url: "postgres://user:password@db.example.com:5432/openfiltr?sslmode=require"
  
  # Custom port
  database_url: "postgres://openfiltr:openfiltr@localhost:5433/openfiltr"
```

**SSL Modes:**
- `disable` - No SSL (development only)
- `require` - SSL enabled
- `verify-full` - SSL with full verification

---

## Authentication Options

### `auth.token_expiry_hours`

| Property | Value |
|----------|-------|
| Type | integer |
| Default | `24` |
| Env Variable | `OPENFILTR_TOKEN_EXPIRY` |

JWT token expiry time in hours.

**Examples:**
```yaml
auth:
  token_expiry_hours: 24      # 1 day
  token_expiry_hours: 168     # 1 week
  token_expiry_hours: 720     # 30 days
```

---

## Environment Variables

All configuration options can be set via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENFILTR_HTTP_PORT` | HTTP listen port | `3000` |
| `OPENFILTR_DNS_PORT` | DNS listen port | `53` |
| `OPENFILTR_BASE_URL` | Public base URL | `http://localhost:3000` |
| `OPENFILTR_DNS_CACHE_TTL` | DNS cache TTL (seconds) | `300` |
| `OPENFILTR_UPSTREAM_DNS` | Comma-separated upstream DNS | `1.1.1.1:53,9.9.9.9:53` |
| `DATABASE_URL` | PostgreSQL connection string | (see above) |
| `OPENFILTR_JWT_SECRET` | JWT signing secret | (required) |
| `OPENFILTR_TOKEN_EXPIRY` | Token expiry (hours) | `24` |

**Important:** Never store secrets in YAML files. Use environment variables:

```bash
# Good
export OPENFILTR_JWT_SECRET="your-secret-key"

# Bad - don't do this!
auth:
  jwt_secret: "your-secret-key"  # DON'T!
```

---

## Complete Example

```yaml
version: 1

server:
  listen_http: ":3000"
  listen_dns: ":53"
  base_url: "https://dns.mydomain.com"

dns:
  cache_ttl: 600
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
    - name: Google
      address: "8.8.8.8:53"

storage:
  database_url: "postgres://user:pass@db.example.com:5432/openfiltr?sslmode=require"

auth:
  token_expiry_hours: 168
```

---

## Related Links

- [Installation Guide](./installation.md)
- [API Reference](./api.md)
- [README](../README.md)
