#!/bin/sh
set -eu

cd "$(dirname "$0")"
action=${1:-start}
if [ "$#" -gt 1 ] || { [ "$action" != start ] && [ "$action" != restart ]; }; then
    echo "Usage: $0 [start|restart]" >&2
    exit 2
fi
if [ ! -f release.env ] || [ ! -f release.config.yaml ]; then
    echo "Create release.env and release.config.yaml first; see docs/release.en.md." >&2
    exit 1
fi
if ! grep -Eq '^[[:space:]]*jwt_key:[[:space:]]*("[[:xdigit:]]{64,}"|[[:xdigit:]]{64,})[[:space:]]*$' release.config.yaml; then
    echo "Set jwt_key to at least 32 random bytes encoded as hex in release.config.yaml; see docs/release.en.md." >&2
    exit 1
fi
if ! command -v docker >/dev/null 2>&1 || ! docker compose version >/dev/null 2>&1; then
    echo "Docker and the Docker Compose plugin are required." >&2
    exit 1
fi

docker compose --env-file release.env -p stocks-calculator-release -f compose.release.yaml up -d --no-build
if [ "$action" = restart ]; then
    exec docker compose --env-file release.env -p stocks-calculator-release -f compose.release.yaml restart server frontend gateway
fi
