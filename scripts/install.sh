#!/usr/bin/env bash
# scripts/install.sh — OpenFiltr one-line installer for Linux
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.sh | sh
#   bash install.sh --version v1.0.0 --dry-run

set -euo pipefail

REPO="openfiltr/openfiltr"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/openfiltr"
DATA_DIR="/var/lib/openfiltr"
SERVICE_USER="openfiltr"
SERVICE_FILE="/etc/systemd/system/openfiltr.service"
BINARY_NAME="openfiltr"
VERSION="${OPENFILTR_VERSION:-latest}"
DRY_RUN=false
NO_ROOT=false

# ── Colours ───────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; CYAN='\033[0;36m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; RESET='\033[0m'
info()    { printf "${CYAN}  →  %s${RESET}\n" "$*"; }
success() { printf "${GREEN}  ✓  %s${RESET}\n" "$*"; }
warn()    { printf "${YELLOW}  !  %s${RESET}\n" "$*"; }
fatal()   { printf "${RED}  ✗  %s${RESET}\n" "$*" >&2; exit 1; }

# ── Argument parsing ──────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case $1 in
    --version)  VERSION="$2"; shift 2 ;;
    --dry-run)  DRY_RUN=true; shift ;;
    --no-root)  NO_ROOT=true; shift ;;
    *) fatal "Unknown argument: $1" ;;
  esac
done

if [[ "$DRY_RUN" == "true" ]]; then
  warn "DRY RUN — no changes will be made"
fi

if [[ "$NO_ROOT" == "true" ]]; then
  INSTALL_DIR="${HOME}/.local/bin"
  CONFIG_DIR="${HOME}/.config/openfiltr"
  DATA_DIR="${HOME}/.local/share/openfiltr"
fi

# ── OS / Arch detection ───────────────────────────────────────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  armv7l)          ARCH="armv7" ;;
  *) fatal "Unsupported architecture: $ARCH" ;;
esac

[[ "$OS" != "linux" ]] && fatal "This installer supports Linux only. For macOS, download the binary from the releases page."

BINARY_FILENAME="${BINARY_NAME}-${OS}-${ARCH}"

# ── Resolve latest version ────────────────────────────────────────────────────
if [[ "$VERSION" == "latest" ]]; then
  info "Resolving latest release…"
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | cut -d'"' -f4)
  [[ -z "$VERSION" ]] && fatal "Could not determine latest release version"
fi

success "Installing OpenFiltr ${VERSION} (${OS}/${ARCH})"

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_FILENAME}"
CHECKSUM_URL="${DOWNLOAD_URL}.sha256"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

# ── Download ──────────────────────────────────────────────────────────────────
info "Downloading binary from ${DOWNLOAD_URL}…"
if [[ "$DRY_RUN" == "false" ]]; then
  curl -fsSL --connect-timeout 30 -o "${TMP_DIR}/${BINARY_FILENAME}" "$DOWNLOAD_URL" \
    || fatal "Download failed"

  # Verify checksum if available
  if curl -fsSL --connect-timeout 10 -o "${TMP_DIR}/${BINARY_FILENAME}.sha256" "$CHECKSUM_URL" 2>/dev/null; then
    cd "$TMP_DIR"
    sha256sum -c "${BINARY_FILENAME}.sha256" --quiet \
      && success "Checksum verified" \
      || fatal "Checksum verification failed — aborting"
    cd - > /dev/null
  else
    warn "Checksum file not found — skipping verification"
  fi

  chmod +x "${TMP_DIR}/${BINARY_FILENAME}"
fi

# ── Install binary ────────────────────────────────────────────────────────────
info "Installing binary to ${INSTALL_DIR}/${BINARY_NAME}…"
if [[ "$DRY_RUN" == "false" ]]; then
  mkdir -p "$INSTALL_DIR"
  if [[ "$NO_ROOT" == "false" ]]; then
    sudo install -m 0755 "${TMP_DIR}/${BINARY_FILENAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  else
    install -m 0755 "${TMP_DIR}/${BINARY_FILENAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  fi
fi
success "Binary installed"

# ── Create directories ────────────────────────────────────────────────────────
info "Creating directories…"
if [[ "$DRY_RUN" == "false" ]]; then
  if [[ "$NO_ROOT" == "false" ]]; then
    sudo mkdir -p "$CONFIG_DIR" "$DATA_DIR"
    sudo useradd -r -s /sbin/nologin -d "$DATA_DIR" "$SERVICE_USER" 2>/dev/null || true
    sudo chown -R "${SERVICE_USER}:${SERVICE_USER}" "$CONFIG_DIR" "$DATA_DIR"
  else
    mkdir -p "$CONFIG_DIR" "$DATA_DIR"
  fi
fi
success "Directories created"

# ── Write default config ──────────────────────────────────────────────────────
CONFIG_FILE="${CONFIG_DIR}/app.yaml"
info "Writing default configuration to ${CONFIG_FILE}…"
if [[ "$DRY_RUN" == "false" ]] && [[ ! -f "$CONFIG_FILE" ]]; then
  DEFAULT_CONFIG="version: 1
server:
  listen_http: \":3000\"
  listen_dns: \":53\"
dns:
  upstream_servers:
    - name: Cloudflare
      address: \"1.1.1.1:53\"
    - name: Quad9
      address: \"9.9.9.9:53\"
storage:
  path: \"${DATA_DIR}/openfiltr.db\"
auth:
  token_expiry_hours: 24
"
  if [[ "$NO_ROOT" == "false" ]]; then
    echo "$DEFAULT_CONFIG" | sudo tee "$CONFIG_FILE" > /dev/null
    sudo chmod 0600 "$CONFIG_FILE"
    sudo chown "${SERVICE_USER}:${SERVICE_USER}" "$CONFIG_FILE"
  else
    echo "$DEFAULT_CONFIG" > "$CONFIG_FILE"
    chmod 0600 "$CONFIG_FILE"
  fi
fi
success "Configuration written"

# ── Write systemd unit ────────────────────────────────────────────────────────
if [[ "$NO_ROOT" == "false" ]] && command -v systemctl &>/dev/null; then
  info "Writing systemd service to ${SERVICE_FILE}…"
  if [[ "$DRY_RUN" == "false" ]]; then
    sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=OpenFiltr DNS Filtering Service
Documentation=https://github.com/openfiltr/openfiltr
After=network.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
ExecStart=${INSTALL_DIR}/${BINARY_NAME} --config ${CONFIG_FILE}
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
EOF
    sudo systemctl daemon-reload
    sudo systemctl enable --now openfiltr
  fi
  success "systemd service enabled and started"
fi

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
printf "${GREEN}"
echo "╔══════════════════════════════════════════════╗"
echo "║      OpenFiltr installed successfully! 🎉    ║"
echo "║                                              ║"
echo "║  Open your browser:  http://localhost:3000   ║"
echo "║  Complete setup to create your admin user.   ║"
echo "╚══════════════════════════════════════════════╝"
printf "${RESET}"
echo ""
info "To view logs: sudo journalctl -u openfiltr -f"
info "To stop:      sudo systemctl stop openfiltr"
