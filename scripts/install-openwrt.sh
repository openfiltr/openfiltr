#!/bin/sh

set -eu

REPO="${OPENFILTR_REPO:-openfiltr/openfiltr}"
VERSION="${OPENFILTR_VERSION:-latest}"
DOWNLOAD_URL="${OPENFILTR_DOWNLOAD_URL:-}"
CHECKSUM_URL="${OPENFILTR_CHECKSUM_URL:-}"
ROUTER_IP=""
HTTP_PORT=""
DNS_PORT=""
DNSMASQ_MODE=""
ASSUME_YES=false

INSTALL_DIR="/usr/bin"
CONFIG_DIR="/etc/openfiltr"
CONFIG_FILE="${CONFIG_DIR}/app.yaml"
SERVICE_FILE="/etc/init.d/openfiltr"
DHCP_CONFIG="/etc/config/dhcp"
DNSMASQ_STATE_FILE="${CONFIG_DIR}/dnsmasq-forward"
ASSET_FILENAME=""
BASE_URL=""
TMP_DIR=""
MODEL_NAME=""

usage() {
  cat <<'EOF'
Usage:
  curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install-openwrt.sh | sh
  sh install-openwrt.sh [options]

Options:
  --version <tag>           Release tag to install (default: latest)
  --repo <owner/repo>       GitHub repository to fetch from
  --router-ip <ip>          Router IP to write into base_url
  --http-port <port>        HTTP listen port (default: 3000)
  --dns-port <port>         DNS listen port (default: 5353)
  --dnsmasq-mode <mode>     'forward' or 'exclusive'
  --download-url <url>      Override the release asset URL
  --checksum-url <url>      Override the checksums URL
  --yes                     Accept detected defaults without prompts
  --help                    Show this help
EOF
}

info() {
  printf '==> %s\n' "$*"
}

warn() {
  printf 'WARNING: %s\n' "$*" >&2
}

fatal() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

cleanup() {
  if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
    rm -rf "$TMP_DIR"
  fi
}

trap cleanup EXIT INT TERM

require_arg() {
  if [ "$#" -lt 2 ] || [ -z "$2" ]; then
    fatal "Missing value for $1"
  fi

  case "$2" in
    -*)
      fatal "Missing value for $1"
      ;;
  esac
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      require_arg "$@"
      VERSION="$2"
      shift 2
      ;;
    --repo)
      require_arg "$@"
      REPO="$2"
      shift 2
      ;;
    --router-ip)
      require_arg "$@"
      ROUTER_IP="$2"
      shift 2
      ;;
    --http-port)
      require_arg "$@"
      HTTP_PORT="$2"
      shift 2
      ;;
    --dns-port)
      require_arg "$@"
      DNS_PORT="$2"
      shift 2
      ;;
    --dnsmasq-mode)
      require_arg "$@"
      DNSMASQ_MODE="$2"
      shift 2
      ;;
    --download-url)
      require_arg "$@"
      DOWNLOAD_URL="$2"
      shift 2
      ;;
    --checksum-url)
      require_arg "$@"
      CHECKSUM_URL="$2"
      shift 2
      ;;
    --yes|-y)
      ASSUME_YES=true
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      fatal "Unknown argument: $1"
      ;;
  esac
done

command -v curl >/dev/null 2>&1 || fatal "curl is required"
command -v tar >/dev/null 2>&1 || fatal "tar is required"
command -v install >/dev/null 2>&1 || fatal "install is required"
command -v uci >/dev/null 2>&1 || fatal "uci is required on OpenWrt"

[ "$(id -u)" -eq 0 ] || fatal "Run this installer as root or with sudo"

ARCH="$(uname -m)"
case "$ARCH" in
  aarch64|arm64)
    ;;
  *)
    fatal "Unsupported architecture: ${ARCH}. MT3000 and MT6000 require linux/arm64 builds."
    ;;
esac

has_tty() {
  [ -r /dev/tty ] && [ -w /dev/tty ]
}

prompt_with_default() {
  prompt_label="$1"
  prompt_default="$2"
  prompt_value="$prompt_default"

  if [ "$ASSUME_YES" = "true" ]; then
    printf '%s' "$prompt_value"
    return 0
  fi

  if has_tty; then
    printf '%s [%s]: ' "$prompt_label" "$prompt_default" > /dev/tty
    if IFS= read -r prompt_input < /dev/tty && [ -n "$prompt_input" ]; then
      prompt_value="$prompt_input"
    fi
  else
    warn "No interactive TTY detected. Using ${prompt_default} for ${prompt_label}."
  fi

  printf '%s' "$prompt_value"
}

confirm_with_default() {
  confirm_label="$1"
  confirm_default="$2"
  confirm_answer="$confirm_default"

  if [ "$ASSUME_YES" = "true" ]; then
    return 0
  fi

  if has_tty; then
    printf '%s [%s]: ' "$confirm_label" "$confirm_default" > /dev/tty
    if IFS= read -r confirm_input < /dev/tty && [ -n "$confirm_input" ]; then
      confirm_answer="$confirm_input"
    fi
  else
    warn "No interactive TTY detected. Using ${confirm_default} for ${confirm_label}."
  fi

  case "$confirm_answer" in
    y|Y|yes|YES)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

validate_port() {
  case "$1" in
    ''|*[!0-9]*)
      return 1
      ;;
  esac

  [ "$1" -ge 1 ] 2>/dev/null && [ "$1" -le 65535 ] 2>/dev/null
}

detect_model() {
  detected_model=""

  if command -v ubus >/dev/null 2>&1 && command -v jsonfilter >/dev/null 2>&1; then
    detected_model="$(ubus call system board 2>/dev/null | jsonfilter -e '@.model' 2>/dev/null || true)"
  fi

  if [ -z "$detected_model" ] && [ -r /tmp/sysinfo/model ]; then
    detected_model="$(cat /tmp/sysinfo/model 2>/dev/null || true)"
  fi

  if [ -z "$detected_model" ] && [ -r /tmp/sysinfo/board_name ]; then
    detected_model="$(cat /tmp/sysinfo/board_name 2>/dev/null || true)"
  fi

  printf '%s' "$detected_model"
}

detect_router_ip() {
  detected_ip="$(uci -q get network.lan.ipaddr 2>/dev/null || true)"

  if [ -z "$detected_ip" ] && command -v ubus >/dev/null 2>&1 && command -v jsonfilter >/dev/null 2>&1; then
    detected_ip="$(ubus call network.interface.lan status 2>/dev/null | jsonfilter -e '@["ipv4-address"][0].address' 2>/dev/null || true)"
  fi

  if [ -z "$detected_ip" ] && command -v ip >/dev/null 2>&1; then
    detected_ip="$(ip -4 addr show br-lan 2>/dev/null | awk '/inet / { sub(/\/.*/, "", $2); print $2; exit }')"
  fi

  if [ -z "$detected_ip" ]; then
    detected_ip="192.168.1.1"
  fi

  printf '%s' "$detected_ip"
}

asset_from_model() {
  model_key="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"

  case "$model_key" in
    *mt3000*|*gl-mt3000*)
      printf '%s' 'openfiltr-openwrt-mt3000.tar.gz'
      ;;
    *mt6000*|*gl-mt6000*)
      printf '%s' 'openfiltr-openwrt-mt6000.tar.gz'
      ;;
    *)
      printf '%s' 'openfiltr-openwrt-arm64.tar.gz'
      ;;
  esac
}

resolve_latest_version() {
  latest_version="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  [ -n "$latest_version" ] || fatal "Could not determine the latest release tag from ${REPO}"
  printf '%s' "$latest_version"
}

backup_if_exists() {
  if [ -e "$1" ]; then
    backup_path="$1.bak.$(date +%Y%m%d%H%M%S)"
    cp "$1" "$backup_path"
    info "Backed up $1 to $backup_path"
  fi
}

read_managed_dnsmasq_server() {
  if [ -r "$DNSMASQ_STATE_FILE" ]; then
    cat "$DNSMASQ_STATE_FILE"
  fi
}

list_dnsmasq_loopback_servers() {
  uci show dhcp 2>/dev/null | sed -n "s/.*server='\\(127\\.0\\.0\\.1#[0-9][0-9]*\\)'.*/\\1/p"
}

warn_unmanaged_dnsmasq_servers() {
  managed_server="$1"
  desired_server="$2"
  existing_servers="$3"

  if [ -n "$existing_servers" ]; then
    for existing_server in $existing_servers; do
      if [ "$existing_server" != "$managed_server" ] && [ "$existing_server" != "$desired_server" ]; then
        warn "Leaving existing dnsmasq loopback server entry ${existing_server} untouched"
      fi
    done
  fi
}

configure_dnsmasq() {
  backup_if_exists "$DHCP_CONFIG"
  managed_server="$(read_managed_dnsmasq_server)"
  existing_servers="$(list_dnsmasq_loopback_servers)"

  if [ "$DNSMASQ_MODE" = "exclusive" ]; then
    if [ -n "$managed_server" ]; then
      uci -q del_list dhcp.@dnsmasq[0].server="$managed_server" || true
    fi
    uci set dhcp.@dnsmasq[0].port='0'
    rm -f "$DNSMASQ_STATE_FILE"
    warn_unmanaged_dnsmasq_servers "$managed_server" "" "$existing_servers"
  else
    desired_server="127.0.0.1#${DNS_PORT}"

    if [ -n "$managed_server" ] && [ "$managed_server" != "$desired_server" ]; then
      uci -q del_list dhcp.@dnsmasq[0].server="$managed_server" || true
    fi

    uci -q delete dhcp.@dnsmasq[0].port || true
    warn_unmanaged_dnsmasq_servers "$managed_server" "$desired_server" "$existing_servers"

    if ! printf '%s\n' "$existing_servers" | grep -qx "$desired_server"; then
      uci add_list dhcp.@dnsmasq[0].server="$desired_server"
    fi

    printf '%s\n' "$desired_server" > "$DNSMASQ_STATE_FILE"
    chmod 0600 "$DNSMASQ_STATE_FILE"
  fi

  uci commit dhcp
  /etc/init.d/dnsmasq restart
}

write_config() {
  umask 077
  mkdir -p "$CONFIG_DIR"
  cat > "$CONFIG_FILE" <<EOF
version: 1

server:
  listen_http: ":${HTTP_PORT}"
  listen_dns: ":${DNS_PORT}"
  base_url: "${BASE_URL}"

storage:
  database_path: "openfiltr.db"

dns:
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
EOF
  chmod 0600 "$CONFIG_FILE"
}

write_service() {
  cat > "$SERVICE_FILE" <<'EOF'
#!/bin/sh /etc/rc.common

USE_PROCD=1
START=95
STOP=10

start_service() {
  procd_open_instance
  procd_set_param command /usr/bin/openfiltr --config /etc/openfiltr/app.yaml
  procd_set_param respawn 3600 5 5
  procd_close_instance
}
EOF
  chmod 0755 "$SERVICE_FILE"
}

verify_health() {
  health_attempt=1
  while [ "$health_attempt" -le 10 ]; do
    if curl -fsS "http://127.0.0.1:${HTTP_PORT}/api/v1/system/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
    health_attempt=$((health_attempt + 1))
  done
  return 1
}

MODEL_NAME="$(detect_model)"
ASSET_FILENAME="$(asset_from_model "$MODEL_NAME")"

if [ -z "$MODEL_NAME" ]; then
  warn "Could not detect the router model. Falling back to the generic OpenWrt arm64 package."
else
  info "Detected router model: ${MODEL_NAME}"
fi

if [ -z "$ROUTER_IP" ]; then
  ROUTER_IP="$(prompt_with_default "Router IP for the API base URL" "$(detect_router_ip)")"
fi

if [ -z "$HTTP_PORT" ]; then
  HTTP_PORT="$(prompt_with_default "HTTP port" "3000")"
fi

if [ -z "$DNS_PORT" ]; then
  DNS_PORT="$(prompt_with_default "DNS port" "5353")"
fi

validate_port "$HTTP_PORT" || fatal "Invalid HTTP port: ${HTTP_PORT}"
validate_port "$DNS_PORT" || fatal "Invalid DNS port: ${DNS_PORT}"

case "$DNSMASQ_MODE" in
  ''|forward|exclusive)
    ;;
  *)
    fatal "Unsupported dnsmasq mode: ${DNSMASQ_MODE}. Use 'forward' or 'exclusive'."
    ;;
esac

if [ "$DNS_PORT" = "53" ]; then
  if [ -z "$DNSMASQ_MODE" ]; then
    if confirm_with_default "OpenFiltr cannot share port 53 with dnsmasq. Disable dnsmasq DNS listening?" "n"; then
      DNSMASQ_MODE="exclusive"
    else
      fatal "Pick a non-53 DNS port or rerun with --dnsmasq-mode exclusive."
    fi
  fi

  [ "$DNSMASQ_MODE" = "exclusive" ] || fatal "dnsmasq mode must be 'exclusive' when DNS port 53 is used."
else
  if [ -z "$DNSMASQ_MODE" ]; then
    DNSMASQ_MODE="forward"
  fi

  [ "$DNSMASQ_MODE" = "forward" ] || fatal "dnsmasq mode 'exclusive' only makes sense with DNS port 53."
fi

BASE_URL="http://${ROUTER_IP}:${HTTP_PORT}"

if [ -z "$DOWNLOAD_URL" ]; then
  if [ "$VERSION" = "latest" ]; then
    info "Resolving the latest release from ${REPO}"
    VERSION="$(resolve_latest_version)"
  fi
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_FILENAME}"
  if [ -z "$CHECKSUM_URL" ]; then
    CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
  fi
else
  ASSET_FILENAME="$(basename "$DOWNLOAD_URL")"
  if [ -z "$CHECKSUM_URL" ]; then
    warn "No checksum URL supplied with --download-url. Checksum verification will be skipped."
  fi
fi

info "Install settings:"
info "  Release: ${VERSION}"
info "  Asset: ${ASSET_FILENAME}"
info "  HTTP listen: :${HTTP_PORT}"
info "  DNS listen: :${DNS_PORT}"
info "  base_url: ${BASE_URL}"
info "  dnsmasq mode: ${DNSMASQ_MODE}"

if [ "$ASSUME_YES" != "true" ]; then
  if ! confirm_with_default "Proceed with installation?" "y"; then
    fatal "Installation cancelled"
  fi
fi

TMP_DIR="$(mktemp -d)"
ARCHIVE_PATH="${TMP_DIR}/${ASSET_FILENAME}"
CHECKSUMS_PATH="${TMP_DIR}/checksums.txt"

info "Downloading ${DOWNLOAD_URL}"
curl -fsSL --connect-timeout 30 -o "$ARCHIVE_PATH" "$DOWNLOAD_URL" || fatal "Failed to download ${DOWNLOAD_URL}"

if [ -n "$CHECKSUM_URL" ]; then
  if curl -fsSL --connect-timeout 15 -o "$CHECKSUMS_PATH" "$CHECKSUM_URL" 2>/dev/null; then
    if command -v sha256sum >/dev/null 2>&1; then
      checksum_line="$(grep " ${ASSET_FILENAME}\$" "$CHECKSUMS_PATH" | head -n 1 || true)"
      if [ -n "$checksum_line" ]; then
        printf '%s\n' "$checksum_line" > "${TMP_DIR}/${ASSET_FILENAME}.sha256"
        (
          cd "$TMP_DIR"
          sha256sum -c "${ASSET_FILENAME}.sha256"
        ) >/dev/null 2>&1 || fatal "Checksum verification failed for ${ASSET_FILENAME}"
        info "Checksum verified"
      else
        warn "No checksum entry found for ${ASSET_FILENAME}. Skipping verification."
      fi
    else
      warn "sha256sum is unavailable. Skipping checksum verification."
    fi
  else
    warn "Could not download ${CHECKSUM_URL}. Skipping checksum verification."
  fi
fi

info "Extracting ${ASSET_FILENAME}"
tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR" || fatal "Failed to extract ${ASSET_FILENAME}"
[ -f "${TMP_DIR}/openfiltr" ] || fatal "The archive did not contain an 'openfiltr' binary"
chmod 0755 "${TMP_DIR}/openfiltr"

if [ -x "$SERVICE_FILE" ]; then
  info "Stopping the existing OpenFiltr service"
  "$SERVICE_FILE" stop >/dev/null 2>&1 || true
fi

backup_if_exists "$CONFIG_FILE"
backup_if_exists "$SERVICE_FILE"

info "Installing the OpenFiltr binary"
install -m 0755 "${TMP_DIR}/openfiltr" "${INSTALL_DIR}/openfiltr"

info "Writing ${CONFIG_FILE}"
write_config

info "Writing ${SERVICE_FILE}"
write_service

if [ "$DNSMASQ_MODE" = "exclusive" ]; then
  info "Updating dnsmasq to free port 53 for OpenFiltr"
  configure_dnsmasq
fi

info "Enabling and starting OpenFiltr"
"$SERVICE_FILE" enable
"$SERVICE_FILE" restart

if [ "$DNSMASQ_MODE" = "forward" ]; then
  info "Updating dnsmasq to forward queries to OpenFiltr on ${DNS_PORT}"
  configure_dnsmasq
fi

if verify_health; then
  info "OpenFiltr is responding on http://127.0.0.1:${HTTP_PORT}/api/v1/system/health"
else
  fatal "OpenFiltr did not pass the local health check. Inspect 'logread -e openfiltr' on the router."
fi

cat <<EOF

OpenFiltr has been installed on this router.

HTTP API: ${BASE_URL}
DNS listen: :${DNS_PORT}
dnsmasq mode: ${DNSMASQ_MODE}

Useful checks:
  logread -e openfiltr
  curl -fsS http://127.0.0.1:${HTTP_PORT}/api/v1/system/health
EOF
