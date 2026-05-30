#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-full}"

run_backend_module_checks() {
  echo "[architecture][backend] module boundaries"
  (
    cd "$ROOT_DIR/code/backend"
    go test ./internal/module -run TestModuleArchitectureBoundaries
  )
}

run_backend_app_architecture_checks() {
  echo "[architecture][backend] app composition boundaries"
  (
    cd "$ROOT_DIR/code/backend"
    go test ./internal/app -run 'TestArchitectureRulesRejectConcreteCrossModuleImports|TestBackendOperationalBoundariesUseContext|TestBackendDoesNotExposeWithContextNames|TestContextBackgroundOnlyAtApprovedRoots'
  )
}

case "$MODE" in
  --quick|quick)
    run_backend_module_checks
    ;;
  --full|full)
    run_backend_module_checks
    run_backend_app_architecture_checks
    ;;
  *)
    echo "usage: scripts/check-backend-architecture.sh [--quick|--full]" >&2
    exit 2
    ;;
esac
