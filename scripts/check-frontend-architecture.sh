#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-full}"

run_frontend_layer_checks() {
  echo "[architecture][frontend] layer boundaries"
  (
    cd "$ROOT_DIR/code/frontend"
    npm run test:run -- \
      src/__tests__/architectureBoundaries.test.ts \
      src/views/__tests__/routeViewArchitectureBoundary.test.ts
  )
}

run_frontend_growth_checks() {
  echo "[architecture][frontend] growth guards"
  (cd "$ROOT_DIR/code/frontend" && npm run check:frontend-growth)
}

run_frontend_feature_boundary_checks() {
  echo "[architecture][frontend] feature owner boundaries"
  (
    cd "$ROOT_DIR/code/frontend"
    npm run test:run -- \
      src/features/contest-awd-admin/model/useAwdOwnerBoundaries.test.ts \
      src/features/contest-awd-admin/model/usePlatformContestAwdBoundary.test.ts \
      src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts
  )
}

run_frontend_overlay_checks() {
  echo "[architecture][frontend] overlay boundaries"
  (
    cd "$ROOT_DIR/code/frontend"
    npm run test:run -- src/components/common/__tests__/ModalTemplates.test.ts
  )
}

run_frontend_theme_checks() {
  echo "[architecture][frontend] theme token boundaries"
  (cd "$ROOT_DIR/code/frontend" && npm run check:theme-tail)
}

case "$MODE" in
  --quick|quick)
    run_frontend_layer_checks
    ;;
  --full|full)
    run_frontend_layer_checks
    run_frontend_growth_checks
    run_frontend_feature_boundary_checks
    run_frontend_overlay_checks
    run_frontend_theme_checks
    ;;
  *)
    echo "usage: scripts/check-frontend-architecture.sh [--quick|--full]" >&2
    exit 2
    ;;
esac
