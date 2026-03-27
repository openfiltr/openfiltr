# Configuration Reference

OpenFiltr is configured via `config/app.yaml`. All values can be overridden with environment variables.

## Example configuration

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

## Options

### `version` *(integer, default: `1`)*

Configuration schema version. Must be `1`.

### `server` *(object)*

| Key | Type | Default | Env Override | Description |
|-----|------|---------|-------------|-------------|
| `listen_http` | string | `":3000"` | `OPENFILTR_LISTEN_HTTP` | HTTP API listen address |
| `listen_dns` | string | `":53"` | `OPENFILTR_LISTEN_DNS` | DNS listen address |
| `base_url` | string | `"http://localhost:3000"` | `OPENFILTR_BASE_URL` | Public URL for callbacks and CSRF |

### `dns` *(object)*

| Key | Type | Default | Env Override | Description |
|-----|------|---------|-------------|-------------|
| `cache_ttl` | integer | `300` | `OPENFILTR_DNS_CACHE_TTL` | DNS response cache TTL in seconds |
| `upstream_servers` | array | — | — | List of upstream DNS resolvers |
| `upstream_servers[].name` | string | — | — | Human-readable server name |
| `upstream_servers[].address` | string | — | — | DNS server address (`host:port`) |

### `storage` *(object)*

| Key | Type | Default | Env Override | Description |
|-----|------|---------|-------------|-------------|
| `database_url` | string | — | `OPENFILTR_DATABASE_URL` | PostgreSQL connection string |

### `auth` *(object)*

| Key | Type | Default | Env Override | Description |
|-----|------|---------|-------------|-------------|
| `token_expiry_hours` | integer | `24` | `OPENFILTR_TOKEN_EXPIRY_HOURS` | JWT token expiry in hours |

> **Security:** Set `OPENFILTR_JWT_SECRET` via environment variable. Never store secrets in YAML config files.

## Environment variables

All YAML keys can be overridden using `OPENFILTR_` prefixed environment variables with uppercase keys and underscores. For example:

- `server.listen_http` → `OPENFILTR_LISTEN_HTTP`
- `dns.cache_ttl` → `OPENFILTR_DNS_CACHE_TTL`
- `storage.database_url` → `OPENFILTR_DATABASE_URL`

The `OPENFILTR_JWT_SECRET` environment variable is **required** for authentication to function. Generate one with:

```bash
openssl rand -hex 32
```
