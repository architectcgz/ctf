#!/usr/bin/env bash
set -euo pipefail

if [ "${BASH_SOURCE[0]}" != "$0" ]; then
  echo "multi-instance nginx proxy smoke must be executed, not sourced" >&2
  return 2
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Keep tools/ as the operator entrypoint; the implementation lives in scripts/lib
# so workflow script-size guards stay focused on public entrypoints.
source "$ROOT_DIR/scripts/lib/multi-instance-nginx-proxy-smoke/run.sh"
