# dnsbro
Modern Go-based DNS daemon for Ubuntu with DoH forwarding, Bubble Tea TUI, and systemd integration.

## Features
- Listens on `127.0.0.1:53` (UDP/TCP) and forwards queries over DNS-over-HTTPS (Cloudflare by default).
- Simple rule engine (allow/block).
- Bubble Tea TUI that streams live queries and stats.
- systemd install/uninstall helpers plus CLI controls (`start`, `stop`, `status`, `reload`).
- Docker integration test that proves real DoH resolution.
- Dev helper script to auto stage/commit/push.

## Quickstart
```bash
go run ./cmd/dnsbro serve --config ./configs/config.yaml
```
Press `Ctrl+C` to exit. Use `--no-tui` if running as a system service.

## Installation
```bash
GOPROXY=direct go install github.com/ogpourya/dnsbro@latest
```
The binary lands in `$GOBIN` (default `$HOME/go/bin`); ensure it is on your `PATH`.

## Configuration
Example file: `configs/config.yaml`
```yaml
listen: 127.0.0.1:53
upstream:
  doh_endpoint: https://1.1.1.1/dns-query
  timeout: 5s
rules:
  blocklist:
    - ads.example.com
  allowlist: []
log:
  file: /var/log/dnsbro.log
  level: info
tui:
  enabled: true
```
- Hot reload: send `SIGHUP` to the process or use `dnsbro reload` when running under systemd.
- If the config file is missing, `dnsbro serve` starts with safe defaults.

## CLI
- `dnsbro serve [--config path] [--no-tui]` – start the daemon (foreground).
- `dnsbro install --config /etc/dnsbro/config.yaml` – install binary + systemd unit.
- `dnsbro uninstall` / `dnsbro revert` – stop/remove service (revert leaves DNS restoration to the user).
- `dnsbro start|stop|status|reload` – thin systemctl wrappers.
- `dnsbro tui` – render the dashboard without starting the daemon (layout smoke test).

## Systemd service
Install requires root:
```bash
sudo ./dnsbro install --config /etc/dnsbro/config.yaml
sudo systemctl status dnsbro
```
The unit uses `ExecStart=/usr/local/bin/dnsbro serve --config /etc/dnsbro/config.yaml --no-tui`.

## Testing DoH end-to-end (Docker)
```bash
test/integration/run.sh
```
This builds a container, runs dnsbro, issues real DNS queries against it, and fails if responses are empty. Requires Docker and outbound network access.

## Unit tests
```bash
GOCACHE=$(pwd)/.cache/go-build go test ./...
```

## Git workflow helper
Use `scripts/dev-commit.sh "feat: describe change"` to auto add/commit/push once a feature is ready. The script assumes `origin` is configured.

## Notes
- DoH is the first upstream type implemented; DoT/DNSCrypt slots can be added by extending `internal/upstream`.
- The TUI subscribes to live query events when `dnsbro serve` runs with UI enabled; the standalone `dnsbro tui` is provided to validate rendering.
