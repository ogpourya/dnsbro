# dnsbro
DNS proxy for Ubuntu: listens on `127.0.0.1:53` (UDP/TCP), forwards over DoH, ships with a Bubble Tea TUI and systemd helpers.

## Quick start (Ubuntu)
1) Install Go 1.21+ and ensure `$GOBIN` is on your `PATH`.  
2) Get dnsbro:  
```bash
GOPROXY=direct go install github.com/ogpourya/dnsbro@latest
```
3) Drop a config (tweak as needed):  
```bash
sudo mkdir -p /etc/dnsbro
sudo cp configs/config.yaml /etc/dnsbro/config.yaml
```
4) Run it in the foreground (shows the TUI by default):  
```bash
dnsbro serve --config /etc/dnsbro/config.yaml
```
5) Point a resolver at it:  
```bash
dig @127.0.0.1 -p 53 example.com
```

Prefer a service?  
```bash
sudo dnsbro install --config /etc/dnsbro/config.yaml --no-tui
sudo systemctl status dnsbro
```
Reload config without restart: `sudo dnsbro reload` (or send `SIGHUP`).

## Config snapshot
```yaml
listen: 127.0.0.1:53
upstream:
  doh_endpoint: https://1.1.1.1/dns-query
  timeout: 5s
rules:
  blocklist: []
  allowlist: []
log:
  file: /var/log/dnsbro.log
  level: info
tui:
  enabled: true
```
- Missing config? `dnsbro serve` falls back to safe defaults.
- Turn off the UI with `--no-tui` when running headless.

## Handy commands
- `dnsbro serve [--config path] [--no-tui]` – run in the foreground.
- `dnsbro install|uninstall|revert` – manage the systemd unit.
- `dnsbro start|stop|status|reload` – systemd wrappers.
- `dnsbro tui` – open the dashboard without starting the daemon.

## Dev + tests
- Unit tests: `GOCACHE=$(pwd)/.cache/go-build go test ./...`
- Integration (Docker, real DoH): `test/integration/run.sh`
