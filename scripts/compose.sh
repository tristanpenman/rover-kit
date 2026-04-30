#!/bin/bash
#
# Bring up the demo docker environment
#
set -euo pipefail

command -v docker >/dev/null 2>&1 || { echo "error: docker is required" >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "error: docker compose v2 is required" >&2; exit 1; }

compose() {
  docker compose --profile demo "$@";
}

echo "==> Building images"
compose build

echo "==> Starting containers"
compose up --remove-orphans -d --force-recreate

echo "==> Tailing logs"
compose logs -f
