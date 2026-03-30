# OpenWrt MT3000 and MT6000 Deployment

This guide assumes a GL.iNet MT3000 or MT6000 running OpenWrt with persistent overlay storage.
OpenFiltr runs as a single binary with the default bbolt backend, so there is no PostgreSQL dependency for this setup.

## Quick Install

Run the installer on the router:

```sh
curl -fsSL https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install-openwrt.sh | sh
```

Run it as `root`, or pipe it into `sudo sh` if your router actually has `sudo`.

The installer:

- Detects MT3000 and MT6000 models and falls back to a generic OpenWrt arm64 asset when detection is unclear
- Downloads the matching GitHub Release archive (`openfiltr-openwrt-mt3000.tar.gz`, `openfiltr-openwrt-mt6000.tar.gz`, or `openfiltr-openwrt-arm64.tar.gz`)
- Prompts for the router IP plus HTTP and DNS ports
- Writes `/etc/openfiltr/app.yaml`
- Installs `/usr/bin/openfiltr`
- Creates `/etc/init.d/openfiltr`
- Configures dnsmasq for forwarding by default

Useful flags:

```sh
sh install-openwrt.sh --version v0.2.0 --router-ip 192.168.8.1 --http-port 3000 --dns-port 5353
sh install-openwrt.sh --dns-port 53 --dnsmasq-mode exclusive
sh install-openwrt.sh --download-url https://example.invalid/openfiltr-openwrt-mt3000.tar.gz --checksum-url https://example.invalid/checksums.txt
```

`--download-url` exists for pre-release or branch testing. The normal install path should pull from GitHub Releases, not ephemeral Actions artefacts.

## Layout

Keep the binary and config under persistent storage:

- Binary: `/usr/bin/openfiltr`
- Config: `/etc/openfiltr/app.yaml`
- Database: `/etc/openfiltr/openfiltr.db`

The database path is relative to the config file in this example, so the bbolt file ends up beside `app.yaml`.

## Manual Fallback

If you do not want the installer, you can still deploy manually with the same layout and config shown below.

## Config

Create `/etc/openfiltr/app.yaml` with a local bbolt store and a DNS listener that does not collide with dnsmasq:

```yaml
version: 1

server:
  listen_http: ":3000"
  listen_dns: ":5353"
  base_url: "http://192.168.1.1:3000"

storage:
  database_path: "openfiltr.db"

dns:
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
```

Adjust `base_url` to the router address you actually use for the API. The installer will prompt for this and default to the detected LAN address.

## procd Service

Create `/etc/init.d/openfiltr`:

```sh
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
```

Then enable and start it:

```sh
chmod +x /etc/init.d/openfiltr
/etc/init.d/openfiltr enable
/etc/init.d/openfiltr start
```

## dnsmasq Guidance

Do not bind both dnsmasq and OpenFiltr to port 53. The installer defaults to forwarding mode for exactly that reason.

The safer option is to keep dnsmasq on 53 and forward queries to OpenFiltr on 5353:

```sh
uci add_list dhcp.@dnsmasq[0].server='127.0.0.1#5353'
uci commit dhcp
/etc/init.d/dnsmasq restart
```

If you insist on putting OpenFiltr on port 53, move dnsmasq first:

```sh
uci set dhcp.@dnsmasq[0].port='0'
uci commit dhcp
/etc/init.d/dnsmasq restart
```

That disables dnsmasq DNS listening on the router. Do it only if you actually want OpenFiltr to own port 53 directly.

## Verification

Check that the service is up and the DNS port is listening:

```sh
logread -e openfiltr
curl -fsS http://127.0.0.1:3000/api/v1/system/health
netstat -lntup | grep 5353
```

If you used the installer, it already ran the local health check before returning.
