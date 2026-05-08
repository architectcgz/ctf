#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-full}"

run_backend_checks() {
  echo "[architecture] backend module boundaries"
  (cd "$ROOT_DIR/code/backend" && go test ./internal/module -run TestModuleArchitectureBoundaries)
}

run_frontend_checks() {
  echo "[architecture] frontend layer boundaries"
  (
    cd "$ROOT_DIR/code/frontend"
    npm run test:run -- \
      src/__tests__/architectureBoundaries.test.ts \
      src/views/__tests__/routeViewArchitectureBoundary.test.ts
  )
}

run_overlay_checks() {
  echo "[architecture] frontend overlay boundaries"
  (
    cd "$ROOT_DIR/code/frontend"
    npm run test:run -- src/components/common/__tests__/ModalTemplates.test.ts
  )
}

run_frontend_theme_checks() {
  echo "[architecture] frontend theme token boundaries"
  (cd "$ROOT_DIR/code/frontend" && npm run check:theme-tail)
}

case "$MODE" in
  --quick|quick)
    run_backend_checks
    run_frontend_checks
    ;;
  --full|full)
    run_backend_checks
    run_frontend_checks
    run_overlay_checks
    run_frontend_theme_checks
    ;;
  *)
    echo "usage: scripts/check-architecture.sh [--quick|--full]" >&2
    exit 2
    ;;
esac
