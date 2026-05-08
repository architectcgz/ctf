#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail=0

red() { printf '\033[31m%s\033[0m' "$1"; }
green() { printf '\033[32m%s\033[0m' "$1"; }

check() {
  local label="$1"
  shift
  if "$@"; then
    echo "  $(green PASS) — $label"
  else
    echo "  $(red FAIL) — $label"
    fail=1
  fi
}

cd "$ROOT_DIR"

echo "[doctor] git hooks"
hooks_path="$(git config --get core.hooksPath || true)"
check "core.hooksPath is .githooks" test "$hooks_path" = ".githooks"
check ".githooks/pre-commit is executable" test -x ".githooks/pre-commit"

echo "[doctor] harness scripts"
check "scripts/check-consistency.sh is executable" test -x "scripts/check-consistency.sh"
check "scripts/check-architecture.sh is executable" test -x "scripts/check-architecture.sh"
check "scripts/doctor-local-harness.sh is executable" test -x "scripts/doctor-local-harness.sh"

echo "[doctor] local tools"
check "git is available" command -v git
check "go is available" command -v go
check "node is available" command -v node
check "npm is available" command -v npm

if [[ -d "code/frontend/node_modules" ]]; then
  echo "  $(green PASS) — code/frontend/node_modules exists"
else
  echo "  $(red FAIL) — code/frontend/node_modules missing; run npm install in code/frontend"
  fail=1
fi

if [[ "$fail" -eq 0 ]]; then
  echo "$(green '✓ local harness doctor passed')"
else
  echo "$(red '✗ local harness doctor found issues')"
fi

exit "$fail"
