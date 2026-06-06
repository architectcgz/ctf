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

echo "[doctor] local tools"
check "git is available" command -v git
check "python3 is available" command -v python3
check "go is available" command -v go
check "node is available" command -v node
check "npm is available" command -v npm

exit "$fail"
