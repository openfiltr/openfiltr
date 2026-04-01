# OpenFiltr Installation Guide

This guide covers all supported installation methods for OpenFiltr. Choose the one that best fits your environment:

- [curl one-liner](#curl-one-liner-linux-macos) — Linux and macOS
- [PowerShell](#powershell-windows) — Windows
- [Docker](#docker) — Container deployment
- [Docker Compose](#docker-compose) — Full stack with persistence
- [Manual binary](#manual-binary-installation) — Direct binary download

---

## curl one-liner (Linux / macOS)

The fastest way to install OpenFiltr on Linux or macOS.

### Prerequisites

- `curl` or `wget`
- `sudo` access (optional, required for system-wide installation)
- One of the following architectures: `amd64`, `arm64`, `armv7`

### Installation

```bash
curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.sh | sh
```

The installer will:

1. Detect your OS and architecture
2. Download the latest release binary
3. Create a systemd service (Linux) or launchd daemon (macOS)
4. Write a default configuration to `/etc/openfiltr/app.yaml`
5. Start the service automatically

#### Customizing the installation

```bash
# Install a specific version
curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.sh | OPENFILTR_VERSION=v0.2.0 sh

# Or with direct script invocation
bash install.sh --version v0.2.0

# Install without root (to ~/.local/bin)
bash install.sh --no-root

# Preview changes without making them
bash install.sh --dry-run
```

### Post-install verification

```bash
# Check service status (Linux)
sudo systemctl status openfiltr

# Check service status (macOS)
sudo launchctl list | grep openfiltr

# Verify the API is responding
curl http://localhost:3000/api/v1/system/health

# View logs (Linux)
sudo journalctl -u openfiltr -f

# View logs (macOS)
tail -f /var/log/openfiltr.log
```

### Raspberry Pi / ARM64

The curl installer fully supports Raspberry Pi and other ARM64 devices. No additional steps are required — the installer auto-detects `arm64` and downloads the correct binary.

```bash
# Standard install works on Raspberry Pi
curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.sh | sh
```

For Raspberry Pi-specific DNS configuration, see the [DNS configuration examples](#dns-configuration-examples) below.

---

## PowerShell (Windows)

Install OpenFiltr on Windows using PowerShell.

### Prerequisites

- PowerShell 5.1 or PowerShell 7+
- Administrator access (optional, required for system-wide installation)

### Installation

Run the following in an **elevated** PowerShell session (Run as Administrator):

```powershell
irm https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.ps1 | iex
```

The installer will:

1. Download the Windows binary
2. Install to `%ProgramFiles%\OpenFiltr\` (or `%LOCALAPPDATA%\OpenFiltr\` with `-NoRoot`)
3. Write default config to `%ProgramData%\openfiltr\app.yaml`
4. Add the binary to your `PATH`
5. Register a Windows service

#### Customizing the installation

```powershell
# Install without elevation (to user profile)
irm https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.ps1 | iex -NoRoot

# Install a specific version
$env:OPENFILTR_VERSION = "v0.2.0"
irm https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.ps1 | iex
```

### Post-install verification

```powershell
# Check service status
Get-Service openfiltr

# Verify the API is responding
Invoke-RestMethod http://localhost:3000/api/v1/system/health

# View logs
Get-WinEvent -LogName Application -FilterXPath "*[System[Provider[@Name='openfiltr']]]" | Select-Object -First 10
```

---

## Docker

Run OpenFiltr in a container.

### Prerequisites

- Docker 20.10+
- Ports 53 (DNS) and 3000 (HTTP API) available

### Quick start

```bash
docker run -d \
  --name openfiltr \
  -p 53:5353/udp \
  -p 53:5353/tcp \
  -p 3000:3000 \
  -v openfiltr-config:/etc/openfiltr \
  --restart unless-stopped \
  ghcr.io/openfiltr/openfiltr:latest
```

> **Note:** The container exposes DNS on port 5353 internally to avoid conflicts with Docker's DNS. Map to port 53 externally.

### ARM64 / Raspberry Pi

The same image supports `amd64` and `arm64`. Docker will pull the correct variant automatically:

```bash
# Works on amd64 and arm64 (Raspberry Pi)
docker run -d \
  --name openfiltr \
  -p 53:5353/udp \
  -p 53:5353/tcp \
  -p 3000:3000 \
  -v openfiltr-config:/etc/openfiltr \
  --restart unless-stopped \
  ghcr.io/openfiltr/openfiltr:latest
```

### Post-install verification

```bash
# Check container status
docker ps | grep openfiltr

# Verify the API is responding
curl http://localhost:3000/api/v1/system/health

# View logs
docker logs openfiltr -f

# Check DNS is listening
docker exec openfiltr netstat -lntup | grep 5353
```

---

## Docker Compose

Recommended for production deployments.

### Prerequisites

- Docker 20.10+
- Docker Compose v2+
- Ports 53 (DNS) and 3000 (HTTP API) available

### Installation

Create a `docker-compose.yml` file:

```yaml
services:
  openfiltr:
    image: ghcr.io/openfiltr/openfiltr:latest
    container_name: openfiltr
    ports:
      - "53:5353/udp"
      - "53:5353/tcp"
      - "3000:3000"
    volumes:
      - openfiltr-config:/etc/openfiltr
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:3000/api/v1/system/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

volumes:
  openfiltr-config:
```

Then run:

```bash
docker compose up -d
```

### With PostgreSQL backend

For larger deployments, you can use PostgreSQL instead of the default bbolt:

```yaml
services:
  openfiltr:
    image: ghcr.io/openfiltr/openfiltr:latest
    container_name: openfiltr
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "53:5353/udp"
      - "53:5353/tcp"
      - "3000:3000"
    volumes:
      - ./app.yaml:/etc/openfiltr/app.yaml:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:3000/api/v1/system/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

  postgres:
    image: postgres:16-alpine
    container_name: openfiltr-postgres
    environment:
      POSTGRES_USER: openfiltr
      POSTGRES_PASSWORD: change-me-in-production
      POSTGRES_DB: openfiltr
    volumes:
      - postgres-data:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U openfiltr"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres-data:
```

Create `app.yaml` in the same directory:

```yaml
version: 1

server:
  listen_http: ":3000"
  listen_dns: ":5353"

storage:
  database_url: "postgres://openfiltr:change-me-in-production@postgres:5432/openfiltr?sslmode=disable"

dns:
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
```

### Post-install verification

```bash
# Check container status
docker compose ps

# Verify the API is responding
curl http://localhost:3000/api/v1/system/health

# View logs
docker compose logs -f openfiltr
```

---

## Manual binary installation

For systems where the automated installers are not suitable.

### Prerequisites

- Go 1.24+ (for building from source), or download a pre-built binary
- Ports 53 (DNS) and 3000 (HTTP API) available

### Download binary

Download the latest release from [GitHub Releases](https://github.com/openfiltr/openfiltr/releases):

```bash
# Linux amd64
curl -L -o openfiltr https://github.com/openfiltr/openfiltr/releases/latest/download/openfiltr-linux-amd64

# Linux arm64 (Raspberry Pi, etc.)
curl -L -o openfiltr https://github.com/openfiltr/openfiltr/releases/latest/download/openfiltr-linux-arm64

# macOS amd64 (Intel)
curl -L -o openfiltr https://github.com/openfiltr/openfiltr/releases/latest/download/openfiltr-darwin-amd64

# macOS arm64 (Apple Silicon)
curl -L -o openfiltr https://github.com/openfiltr/openfiltr/releases/latest/download/openfiltr-darwin-arm64

# Windows amd64
curl -L -o openfiltr.exe https://github.com/openfiltr/openfiltr/releases/latest/download/openfiltr-windows-amd64.exe
```

Make the binary executable (Linux/macOS):

```bash
chmod +x openfiltr
sudo mv openfiltr /usr/local/bin/
```

### Create configuration

Create `/etc/openfiltr/app.yaml`:

```bash
sudo mkdir -p /etc/openfiltr
sudo tee /etc/openfiltr/app.yaml > /dev/null <<EOF
version: 1

server:
  listen_http: ":3000"
  listen_dns: ":53"

storage:
  database_path: "openfiltr.db"

dns:
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
EOF
```

### Run manually

```bash
openfiltr --config /etc/openfiltr/app.yaml
```

### Create systemd service (Linux)

Create `/etc/systemd/system/openfiltr.service`:

```ini
[Unit]
Description=OpenFiltr DNS Filtering Service
Documentation=https://github.com/openfiltr/openfiltr
After=network.target

[Service]
Type=simple
User=openfiltr
Group=openfiltr
ExecStart=/usr/local/bin/openfiltr --config /etc/openfiltr/app.yaml
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
ProtectSystem=strict
PrivateTmp=true
NoNewPrivileges=true
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
# Create user
sudo useradd -r -s /bin/false openfiltr

# Set permissions
sudo chown -R openfiltr:openfiltr /etc/openfiltr

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable --now openfiltr
```

### Post-install verification

```bash
# Check service status
sudo systemctl status openfiltr

# Verify the API is responding
curl http://localhost:3000/api/v1/system/health

# Verify DNS is listening
dig @localhost google.com

# View logs
sudo journalctl -u openfiltr -f
```

---

## DNS configuration examples

After installing OpenFiltr, configure your devices or router to use it as the DNS server.

### Router configuration

#### OpenWrt

Forward DNS queries to OpenFiltr:

```bash
uci add_list dhcp.@dnsmasq[0].server='127.0.0.1#5353'
uci commit dhcp
/etc/init.d/dnsmasq restart
```

Or make OpenFiltr the primary DNS (disable dnsmasq DNS):

```bash
uci set dhcp.@dnsmasq[0].port='0'
uci commit dhcp
/etc/init.d/dnsmasq restart
```

#### pfSense / OPNsense

1. Go to **Services → DNS Resolver**
2. Scroll to **DNS Query Forwarding**
3. Add your OpenFiltr server IP

#### Ubiquiti UniFi

1. Go to **Settings → Networks**
2. Edit your LAN network
3. Set **DNS Server** to your OpenFiltr server IP

### Individual devices

#### Linux

Edit `/etc/resolv.conf`:

```bash
# Replace with your OpenFiltr server IP
nameserver 192.168.1.10
```

Or use NetworkManager:

```bash
nmcli con mod <connection-name> ipv4.dns "192.168.1.10"
nmcli con up <connection-name>
```

#### macOS

**System Settings → Network → [Your connection] → Details → DNS**

Or via command line:

```bash
sudo networksetup -setdnsservers "Wi-Fi" 192.168.1.10
```

#### Windows

**Settings → Network & Internet → [Your connection] → Edit DNS servers**

Or via PowerShell:

```powershell
Set-DnsClientServerAddress -InterfaceAlias "Ethernet" -ServerAddresses 192.168.1.10
```

---

## Troubleshooting

### Port 53 is already in use

**Symptoms:**

- OpenFiltr fails to start
- Error: `bind: address already in use`

**Solution:**

Check what's using port 53:

```bash
sudo lsof -i :53
# or
sudo netstat -lntup | grep :53
```

Common causes:

1. **systemd-resolved** (Ubuntu/Debian)
   ```bash
   sudo systemctl disable --now systemd-resolved
   ```

2. **dnsmasq**
   ```bash
   sudo systemctl stop dnsmasq
   # Or reconfigure to use a different port
   ```

3. **bind9 / named**
   ```bash
   sudo systemctl stop named
   ```

**Alternative:** Configure OpenFiltr to listen on a different port and forward queries:

```yaml
server:
  listen_dns: ":5353"  # Use port 5353 instead of 53
```

Then configure your router or dnsmasq to forward DNS queries to `127.0.0.1#5353`.

### Firewall blocking DNS

**Symptoms:**

- OpenFiltr is running but clients can't resolve DNS
- Timeouts when querying from other devices

**Solution:**

Allow DNS traffic through your firewall:

```bash
# UFW (Ubuntu/Debian)
sudo ufw allow 53/udp
sudo ufw allow 53/tcp

# firewalld (CentOS/RHEL/Fedora)
sudo firewall-cmd --permanent --add-service=dns
sudo firewall-cmd --reload

# iptables
sudo iptables -A INPUT -p udp --dport 53 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 53 -j ACCEPT
```

### Permission denied on port 53

**Symptoms:**

- Error: `bind: permission denied`

**Solution:**

On Linux, non-root processes cannot bind to ports below 1024 without capabilities:

```bash
# Grant capability to bind to privileged ports
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/openfiltr

# Or run as root/sudo
sudo openfiltr --config /etc/openfiltr/app.yaml
```

The systemd service file includes `AmbientCapabilities=CAP_NET_BIND_SERVICE` to handle this automatically.

### Cannot connect to API on port 3000

**Symptoms:**

- `curl http://localhost:3000/api/v1/system/health` fails
- Connection refused or timeout

**Solution:**

1. Check if OpenFiltr is running:
   ```bash
   sudo systemctl status openfiltr
   ```

2. Check logs for errors:
   ```bash
   sudo journalctl -u openfiltr -n 50
   ```

3. Verify the HTTP port is listening:
   ```bash
   sudo netstat -lntup | grep 3000
   ```

4. Check firewall:
   ```bash
   sudo ufw status
   ```

### Database permission errors

**Symptoms:**

- Error: `openfiltr.db: permission denied`
- Service fails to start after config change

**Solution:**

Ensure the OpenFiltr user owns the data directory:

```bash
sudo chown -R openfiltr:openfiltr /etc/openfiltr
sudo chmod 750 /etc/openfiltr
```

For PostgreSQL, verify the connection string and credentials:

```bash
psql "postgres://openfiltr:password@localhost:5432/openfiltr?sslmode=disable"
```

### DNS queries not being blocked

**Symptoms:**

- OpenFiltr is running
- Ads/trackers still resolving

**Solution:**

1. Verify clients are using OpenFiltr as their DNS server:
   ```bash
   # On the client device
   dig google.com
   # Check the SERVER line
   ```

2. Check upstream DNS is working:
   ```bash
   dig @1.1.1.1 google.com
   ```

3. Verify block rules are configured:
   ```bash
   curl -H "Authorization: Bearer <token>" \
     http://localhost:3000/api/v1/filtering/block-rules
   ```

4. Check the logs for query processing:
   ```bash
   sudo journalctl -u openfiltr -f | grep -i query
   ```

### High memory usage

**Symptoms:**

- OpenFiltr consuming more memory than expected

**Solution:**

The default bbolt backend keeps the entire database in memory for performance. For systems with limited RAM:

1. Consider PostgreSQL backend for large rule sets
2. Reduce rule sources
3. Set resource limits (Docker):
   ```yaml
   services:
     openfiltr:
       # ...
       deploy:
         resources:
           limits:
             memory: 512M
   ```

---

## Next steps

After installation:

1. **Create an admin user** — Open `http://localhost:3000` in your browser
2. **Add block rules** — Configure via API or UI
3. **Configure DNS forwarding** — Point your router or devices to OpenFiltr
4. **Review the API docs** — OpenAPI spec at `/openapi.yaml`

For more help:

- [GitHub Issues](https://github.com/openfiltr/openfiltr/issues)
- [Contributing Guide](../CONTRIBUTING.md)
- [Security Policy](../SECURITY.md)