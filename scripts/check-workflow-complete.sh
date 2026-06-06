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
backend_architecture_pattern='(^docs/architecture/|^scripts/check-architecture\.sh$|^scripts/check-backend-architecture\.sh$|^code/backend/)'
frontend_architecture_pattern='(^docs/architecture/|^scripts/check-architecture\.sh$|^scripts/check-frontend-architecture\.sh$|^code/frontend/package\.json$|^code/frontend/scripts/|^code/frontend/src/|^code/frontend/vite\.config\.[^/]+$)'

run_check "review governance" bash scripts/check-review-governance.sh
run_check "code change contract checks" bash scripts/check-code-changes.sh

if [[ -n "$changed" ]] && matches_any "$backend_architecture_pattern" "$changed"; then
  run_check "backend architecture checks" bash scripts/check-backend-architecture.sh --full
else
  echo "[workflow-complete] backend architecture checks"
  echo "  $(green PASS) — no backend architecture-sensitive changes detected"
fi

if [[ -n "$changed" ]] && matches_any "$frontend_architecture_pattern" "$changed"; then
  run_check "frontend architecture checks" bash scripts/check-frontend-architecture.sh --full
else
  echo "[workflow-complete] frontend architecture checks"
  echo "  $(green PASS) — no frontend architecture-sensitive changes detected"
fi

if [[ "$fail" -eq 0 ]]; then
  echo "$(green '✓ workflow completion checks passed')"
else
  echo "$(red '✗ workflow completion checks failed')"
fi

exit "$fail"
