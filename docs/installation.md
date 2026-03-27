# Installation Guide

OpenFiltr can be installed using several methods. Choose the one that best fits your environment.

## Prerequisites

- **Go 1.24+** (only when building from source)
- **PostgreSQL 14+** with a reachable connection string
- **Linux** (x86_64 / ARM64), **macOS**, or **Windows**

> PostgreSQL must be running and accessible before starting OpenFiltr. The connection string is configured in `config/app.yaml` or via the `OPENFILTR_DATABASE_URL` environment variable.

## Method 1: curl (Linux / macOS)

The installer detects your OS and architecture, downloads the binary, sets up a systemd/launchd service, and writes a default configuration.

```bash
curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.sh | sh
```

### Post-install verification

```bash
openfiltr version
systemctl status openfiltr  # Linux
```

## Method 2: Docker

```bash
docker run -d \
  --name openfiltr \
  -p 3000:3000 \
  -p 53:53/udp \
  -e OPENFILTR_DATABASE_URL="postgres://openfiltr:openfiltr@host.docker.internal:5432/openfiltr?sslmode=disable" \
  -e OPENFILTR_JWT_SECRET="change-me" \
  ghcr.io/openfiltr/openfiltr:latest
```

### Post-install verification

```bash
docker logs openfiltr   # check for startup errors
curl http://localhost:3000/api/v1/health
```

## Method 3: Docker Compose

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: openfiltr
      POSTGRES_PASSWORD: openfiltr
      POSTGRES_DB: openfiltr
    volumes:
      - pgdata:/var/lib/postgresql/data

  openfiltr:
    image: ghcr.io/openfiltr/openfiltr:latest
    depends_on: [postgres]
    ports:
      - "3000:3000"
      - "53:53/udp"
    environment:
      OPENFILTR_DATABASE_URL: postgres://openfiltr:openfiltr@postgres:5432/openfiltr?sslmode=disable
      OPENFILTR_JWT_SECRET: change-me

volumes:
  pgdata:
```

```bash
docker compose up -d
curl http://localhost:3000/api/v1/health
```

## Method 4: Manual binary

1. Download the latest release from [GitHub Releases](https://github.com/openfiltr/openfiltr/releases):
   ```bash
   # Linux x86_64
   curl -sL https://github.com/openfiltr/openfiltr/releases/latest/download/openfiltr-linux-amd64 -o /usr/local/bin/openfiltr
   chmod +x /usr/local/bin/openfiltr
   ```

2. Create the config directory:
   ```bash
   mkdir -p /etc/openfiltr
   cp config/app.yaml.example /etc/openfiltr/app.yaml
   ```

3. Edit `/etc/openfiltr/app.yaml` to set your `database_url` and JWT secret.

4. Run:
   ```bash
   openfiltr serve --config /etc/openfiltr/app.yaml
   ```

## Raspberry Pi / ARM64

All installation methods support ARM64. When using Docker, the `linux/arm64` image is automatically selected. For manual installation, download the `linux-arm64` binary from the releases page.

## Troubleshooting

### Port 53 is already in use

On Linux, systemd-resolved often binds port 53 by default:

```bash
# Check what's using port 53
sudo lsof -i :53
sudo ss -ulnp | grep 53

# Release the port (temporary)
sudo systemctl stop systemd-resolved
```

### Firewall rules

Ensure your firewall allows DNS (UDP 53) and API (TCP 3000):

```bash
# UFW
sudo ufw allow 53/udp
sudo ufw allow 3000/tcp

# iptables
sudo iptables -A INPUT -p udp --dport 53 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 3000 -j ACCEPT
```

### Database connection refused

Verify PostgreSQL is running and accepting connections:

```bash
psql "$OPENFILTR_DATABASE_URL" -c "SELECT 1"
```
