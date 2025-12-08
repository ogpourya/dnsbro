# dnsbro
DNS proxy for Ubuntu: listens on `127.0.0.1:53` (UDP/TCP), forwards over DoH, ships with systemd helpers.

## Quick start (Ubuntu)
1) Install Go 1.21+ and ensure `$GOBIN` is on your `PATH`.  
2) Get dnsbro:  
```bash
GOPROXY=direct go install github.com/ogpourya/dnsbro@latest
```
3) Drop a config (tweak as needed):  
```bash
sudo mkdir -p /etc/dnsbro
dnsbro sample-config | sudo tee /etc/dnsbro/config.yaml >/dev/null
```
4) Run it in the foreground:  
```bash
sudo dnsbro serve --config /etc/dnsbro/config.yaml
```
5) Point a resolver at it:  
```bash
dig @127.0.0.1 -p 53 example.com
```

Prefer a service?  
```bash
sudo dnsbro install --config /etc/dnsbro/config.yaml
sudo systemctl status dnsbro
```
Reload config without restart: `sudo dnsbro reload` (or send `SIGHUP`).

## Config snapshot
```yaml
listen: 127.0.0.1:53
upstream:
  doh_endpoint: https://1.1.1.1/dns-query
  timeout: 5s
  bootstrap:
    - 1.1.1.1:53
    - 8.8.8.8:53
rules:
  blocklist: []
  allowlist: []
log:
  file: /var/log/dnsbro.log
  level: info
```
- Missing config? `dnsbro serve` falls back to safe defaults.

## Handy commands
- `dnsbro serve [--config path] [--listen host:port]` – run in the foreground.
- `dnsbro install|uninstall|revert` – manage the systemd unit.
- `dnsbro start|stop|status|reload` – systemd wrappers.
- `dnsbro sample-config` – print the bundled config template.

## Dev + tests
- Unit tests: `GOCACHE=$(pwd)/.cache/go-build go test ./...`
- Integration (Docker, real DoH): `test/integration/run.sh`
