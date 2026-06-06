#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
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

echo "[doctor] local workflow assets"
check ".harness/session-gates directory is available or can be created" test -d ".harness/session-gates" -o ! -e ".harness/session-gates"
check "legacy .harness/reuse-decision.md is absent" test ! -f ".harness/reuse-decision.md"
check "legacy scripts/archive-task-artifacts.sh is absent" test ! -f "scripts/archive-task-artifacts.sh"
check ".gitignore ignores .harness/reuse-index/" grep -qx '/.harness/reuse-index/' ".gitignore"
check ".gitignore ignores .harness/session-gates/" grep -qx '/.harness/session-gates/' ".gitignore"

if bash tools/ensure-frontend-tooling.sh --quiet >/dev/null 2>&1; then
  echo "  $(green PASS) — frontend tooling is available"
else
  echo "  $(red FAIL) — frontend tooling is unavailable; run npm install in code/frontend or ensure main worktree dependencies exist"
  fail=1
fi

exit "$fail"
