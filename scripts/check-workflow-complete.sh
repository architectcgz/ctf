#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail=0

red() { printf '\033[31m%s\033[0m' "$1"; }
green() { printf '\033[32m%s\033[0m' "$1"; }

changed_files() {
  local staged
  staged="$(git diff --cached --name-only)"
  if [[ -n "$staged" ]]; then
    printf '%s\n' "$staged" | sort -u
    return
  fi

  git diff --name-only | sort -u
}

matches_any() {
  local pattern="$1"
  local input="$2"
  if command -v rg >/dev/null 2>&1; then
    printf '%s\n' "$input" | rg -q "$pattern"
  else
    printf '%s\n' "$input" | grep -Eq "$pattern"
  fi
}

run_check() {
  local label="$1"
  shift

  echo "[workflow-complete] $label"
  if "$@"; then
    echo "  $(green PASS) — $label"
  else
    echo "  $(red FAIL) — $label"
    fail=1
  fi
}

cd "$ROOT_DIR"

changed="$(changed_files)"
architecture_pattern='(^docs/architecture/|^scripts/check-architecture\.sh$|^code/backend/internal/module/|^code/frontend/src/__tests__/architectureBoundaries\.test\.ts$|^code/frontend/src/views/__tests__/routeViewArchitectureBoundary\.test\.ts$|^code/frontend/src/components/common/__tests__/ModalTemplates\.test\.ts$|^code/frontend/scripts/check-theme-tail\.mjs$)'

run_check "harness consistency" bash scripts/check-consistency.sh
run_check "code change contract checks" bash scripts/check-code-changes.sh

if [[ -n "$changed" ]] && matches_any "$architecture_pattern" "$changed"; then
  run_check "full architecture checks" bash scripts/check-architecture.sh --full
else
  echo "[workflow-complete] full architecture checks"
  echo "  $(green PASS) — no architecture-sensitive changes detected"
fi

if [[ "$fail" -eq 0 ]]; then
  echo "$(green '✓ workflow completion checks passed')"
else
  echo "$(red '✗ workflow completion checks failed')"
fi

exit "$fail"
