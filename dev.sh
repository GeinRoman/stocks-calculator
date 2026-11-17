#!/bin/sh
set -eu

cd "$(dirname "$0")"
exec docker compose -p stocks-calculator-dev -f compose.dev.yaml up --build "$@"
