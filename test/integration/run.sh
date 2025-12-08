#!/usr/bin/env bash
set -euo pipefail
IMAGE=dnsbro-integration
CONTAINER=dnsbro-integration

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required for integration tests" >&2
  exit 1
fi

echo "[+] building image" 
docker build -t "$IMAGE" -f test/integration/Dockerfile .

echo "[+] starting container"
docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
docker run -d --name "$CONTAINER" -p 1053:53/udp -p 1053:53/tcp "$IMAGE"

cleanup() {
  docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT

# give the daemon a moment to start
sleep 3

echo "[+] querying example.com"
example_ip=$(docker exec "$CONTAINER" dig @127.0.0.1 -p 53 example.com +short | head -n1)
if [[ -z "$example_ip" ]]; then
  echo "example.com did not resolve via dnsbro" >&2
  exit 1
fi

echo "[+] querying google.com"
google_ip=$(docker exec "$CONTAINER" dig @127.0.0.1 -p 53 google.com +short | head -n1)
if [[ -z "$google_ip" ]]; then
  echo "google.com did not resolve via dnsbro" >&2
  exit 1
fi

echo "example.com -> $example_ip"
echo "google.com -> $google_ip"

echo "[+] verifying log output"
docker logs "$CONTAINER" | tail -n 5

echo "[✓] integration test passed"
