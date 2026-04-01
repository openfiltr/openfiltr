<div align="center">

<img src="https://raw.githubusercontent.com/openfiltr/.github/main/assets/logo.svg" alt="OpenFiltr logo" width="120" />

# OpenFiltr

**An open, community-governed DNS filtering backend with API-first design, portable YAML configuration, and self-hosted deployment.**

[![Licence: AGPLv3](https://img.shields.io/badge/Licence-AGPLv3-7C3AED.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](https://go.dev)
[![Build](https://github.com/openfiltr/openfiltr/actions/workflows/ci.yml/badge.svg)](https://github.com/openfiltr/openfiltr/actions)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-green.svg)](openapi/openapi.yaml)

> ⚠️ OpenFiltr is in active development. The API and configuration format are not yet stable.  
> This project was bootstrapped with the assistance of AI. See [CONTRIBUTING.md](CONTRIBUTING.md).

</div>

---

## What is OpenFiltr?

OpenFiltr is a self-hosted DNS filtering server that blocks advertisements, trackers, and malicious domains across your entire network. The repository is currently backend-first. The API, DNS engine, auth, config, storage, Docker, and workflow files are the source of truth. It is built to be:

- **API-first**: every feature is accessible through a documented REST API
- **Portable**: configuration is YAML, importable and exportable in full or by section
- **Fast**: single Go binary, bbolt-backed state by default, PostgreSQL remains optional, low memory footprint
- **Open**: AGPLv3 licence, public roadmap, community governance

## Why does it exist?

Existing DNS filtering tools prioritise simplicity over extensibility. OpenFiltr is designed from the ground up for:

- Operators who want to automate everything through an API
- Teams who want to version-control their DNS filtering configuration
- Developers who want a platform they can extend without forking

## Quick install

> **📖 For detailed installation instructions, prerequisites, and troubleshooting, see the [Installation Guide](docs/installation.md).**

### curl (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.sh | sh
```

The installer detects your OS and architecture, installs a single binary, creates a systemd service (Linux) or launchd daemon (macOS), and writes a default config. The default backend is bbolt, stored beside the config file. Set `storage.database_url` only if you deliberately want PostgreSQL.

The Linux systemd unit is checked in at [`deploy/systemd/openfiltr.service`](deploy/systemd/openfiltr.service), and the installer renders that file with the chosen install and config paths.

### PowerShell (Windows)

Run the following in an **elevated** PowerShell session:

```powershell
irm https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.ps1 | iex
```

The installer downloads the Windows binary, writes a default config to `%ProgramData%\openfiltr\`, adds the binary to your `PATH`, and registers a Windows service. The default backend is bbolt, stored beside the config file. Set `storage.database_url` only if you deliberately want PostgreSQL.

> **Note:** Run PowerShell as Administrator to register the Windows service and write to `%ProgramFiles%`. Pass `-NoRoot` to install to your user profile instead (`%LOCALAPPDATA%\OpenFiltr\`) without requiring elevation.

### Docker Compose

```yaml
services:
  openfiltr:
    image: ghcr.io/openfiltr/openfiltr:latest
    container_name: openfiltr
    ports:
      - "53:53/udp"
      - "53:53/tcp"
      - "3000:3000"
    volumes:
      - openfiltr-config:/etc/openfiltr
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:3000/api/v1/system/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  openfiltr-config:
```

The config volume stores both `app.yaml` and the default bbolt database. Set `storage.database_url` only if you deliberately want PostgreSQL.

### OpenWrt MT3000 / MT6000

Run the router installer on the device itself:

```sh
curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install-openwrt.sh | sh
```

Run it as `root`, or pipe it into `sudo sh` if your router actually has `sudo`.

The installer detects MT3000 and MT6000 models, downloads the matching `linux/arm64` release asset, prompts for router IP plus HTTP and DNS ports, installs `/usr/bin/openfiltr`, writes `/etc/openfiltr/app.yaml`, creates a `procd` service, and updates dnsmasq. The default path keeps dnsmasq on port 53 and forwards to OpenFiltr on 5353. If you choose port 53 for OpenFiltr, the installer disables dnsmasq DNS listening because both processes cannot own the same port.

This path assumes the router has outbound internet access to GitHub release assets and raw GitHub content.

For a router-specific deployment guide, manual fallback steps, and override flags such as `--download-url`, see [the OpenWrt deployment guide](docs/deployment/openwrt-mt3000.md).

## Features

| Feature | v1.0 | v1.1 | v1.2 |
|---|---|---|---|
| DNS forwarding | ✅ | | |
| Block rules (exact, wildcard, regex) | ✅ | | |
| Allow-rule precedence | | planned | |
| Rule sources (hosts, domain lists) | ✅ | | |
| Local DNS entries (A, AAAA, CNAME) | ✅ | | |
| Per-client / per-group policies | | planned | |
| REST API + OpenAPI docs | ✅ | | |
| YAML import & export | ✅ | | |
| Docker & curl install | ✅ | | |
| Embedded bbolt persistence | ✅ | | |
| PostgreSQL persistence | ✅ | | |
| Auth with local users + API tokens | ✅ | | |
| Activity log & audit trail | ✅ | | |
| Configuration export & import | ✅ | | |
| Web UI | | planned | |
| DNSSEC | | ✅ | |
| DoH & DoT upstream support | | ✅ | |
| Webhook events | | ✅ | |
| Prometheus metrics | | ✅ | |
| Role-based access control | | ✅ | |
| Plugin system | | | ✅ |
| SSO | | | ✅ |

## API

The REST API is the primary interface today. Do not assume a working React UI exists in this repository.

```bash
# Get system status
curl -H "Authorization: Bearer <token>" http://localhost:3000/api/v1/system/status

# Add a block rule
curl -X POST http://localhost:3000/api/v1/filtering/block-rules \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"pattern": "*.ads.example.com", "rule_type": "wildcard"}'
```

The OpenAPI 3.1 specification is served at `/openapi.yaml` and documents all endpoints.

## Configuration

OpenFiltr is configured through YAML files in `config/`:

```yaml
version: 1

server:
  listen_http: ":3000"
  listen_dns: ":53"

storage:
  database_path: "openfiltr.db"
  # database_url: "postgres://openfiltr:openfiltr@localhost:5432/openfiltr?sslmode=disable"

dns:
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
```

Relative `database_path` values are resolved against the config file directory, so `openfiltr.db` sits beside `app.yaml`.

Export your configuration at any time:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:3000/api/v1/config/export > config-backup.yaml
```

Exported backups and config bundles start with a schema version header so imports can reject incompatible files clearly:

```yaml
version: 1
block_rules: []
allow_rules: []
rule_sources: []
dns_entries: []
upstream_servers: []
```

## Repo layout

```
/cmd/server       - application entrypoint
/internal         - server internals (not importable)
/openapi          - OpenAPI 3.1 specification
/docs             - documentation
/deploy/docker    - Dockerfile and Compose files
/deploy/systemd   - checked-in systemd unit templates
/scripts          - install.sh, install-openwrt.sh, and helper scripts
/examples         - example configurations
.github/          - CI, templates, issue infrastructure
```

## Contributing

We welcome contributions of all kinds. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a pull request.

All contributors must sign off their commits under the [Developer Certificate of Origin](https://developercertificate.org).

## Security

To report a vulnerability, please use [GitHub Security Advisories](https://github.com/openfiltr/openfiltr/security/advisories/new). **Do not open a public issue.**

See [SECURITY.md](SECURITY.md) for our full disclosure policy.

## Licence

- **Server**: [GNU Affero General Public Licence v3](LICENSE)
- **Documentation**: CC BY 4.0

## Governance

OpenFiltr follows an open governance model. See [GOVERNANCE.md](GOVERNANCE.md) for details on decision-making, the RFC process, and maintainer responsibilities.

See [ROADMAP.md](ROADMAP.md) for the public roadmap.
