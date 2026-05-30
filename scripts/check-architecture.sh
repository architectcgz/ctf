#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-full}"

run_backend_checks() {
  bash "$ROOT_DIR/scripts/check-backend-architecture.sh" "$1"
}

run_frontend_checks() {
  bash "$ROOT_DIR/scripts/check-frontend-architecture.sh" "$1"
}

case "$MODE" in
  --quick|quick)
    run_backend_checks --quick
    run_frontend_checks --quick
    ;;
  --full|full)
    run_backend_checks --full
    run_frontend_checks --full
    ;;
  *)
    echo "usage: scripts/check-architecture.sh [--quick|--full]" >&2
    exit 2
    ;;
esac
