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
- **Fast**: single Go binary, PostgreSQL-backed state, low memory footprint
- **Open**: AGPLv3 licence, public roadmap, community governance

## Why does it exist?

Existing DNS filtering tools prioritise simplicity over extensibility. OpenFiltr is designed from the ground up for:

- Operators who want to automate everything through an API
- Teams who want to version-control their DNS filtering configuration
- Developers who want a platform they can extend without forking

## Quick install

### curl (Linux / Raspberry Pi)

```bash
curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.sh | sh
```

The installer detects your OS and architecture, installs a single binary, creates a systemd service, and writes a default config. PostgreSQL must already be running and reachable from the configured `database_url`.

### Docker Compose

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: openfiltr
      POSTGRES_USER: openfiltr
      POSTGRES_PASSWORD: openfiltr

  openfiltr:
    image: ghcr.io/openfiltr/openfiltr:latest
    depends_on:
      - postgres
    ports:
      - "53:53/udp"
      - "53:53/tcp"
      - "3000:3000"
    environment:
      - OPENFILTR_DATABASE_URL=postgres://openfiltr:openfiltr@postgres:5432/openfiltr?sslmode=disable
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:3000/api/v1/system/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

Then use the API on [http://localhost:3000](http://localhost:3000) to verify the service and complete first-run setup through the auth endpoints.

## Features

| Feature | v1.0 | v1.1 | v1.2 |
|---|---|---|---|
| DNS forwarding | ✅ | | |
| Block & allow rules | ✅ | | |
| Rule sources (hosts, domain lists) | ✅ | | |
| Local DNS entries | ✅ | | |
| Per-client / per-group policies | ✅ | | |
| REST API + OpenAPI docs | ✅ | | |
| YAML import & export | ✅ | | |
| Docker & curl install | ✅ | | |
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
| PostgreSQL | | | ✅ |
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
  database_url: "postgres://openfiltr:openfiltr@localhost:5432/openfiltr?sslmode=disable"

dns:
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
```

Export your configuration at any time:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:3000/api/v1/config/export > config-backup.yaml
```

## Repo layout

```
/cmd/server       - application entrypoint
/internal         - server internals (not importable)
/openapi          - OpenAPI 3.1 specification
/docs             - documentation
/deploy/docker    - Dockerfile and Compose files
/scripts          - install.sh and helper scripts
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
