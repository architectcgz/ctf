#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail=0

red() { printf '\033[31m%s\033[0m' "$1"; }
green() { printf '\033[32m%s\033[0m' "$1"; }

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

run_check "workflow governance stage" bash scripts/run-workflow-stage.sh workflow-governance
run_check "completion-full stage" bash scripts/run-workflow-stage.sh completion-full

if [[ "$fail" -eq 0 ]]; then
  echo "$(green '✓ workflow completion checks passed')"
else
  echo "$(red '✗ workflow completion checks failed')"
fi

exit "$fail"
