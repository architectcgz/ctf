#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail=0

red() { printf '\033[31m%s\033[0m' "$1"; }
green() { printf '\033[32m%s\033[0m' "$1"; }

run_check() {
  local label="$1"
  shift

  if "$@"; then
    echo "$(green "✓ $label passed")"
  else
    echo "$(red "✗ $label failed")"
    fail=1
  fi
}

cd "$ROOT_DIR"

run_check "local harness setup" bash harness/checks/check_local_harness_setup.sh
run_check "local toolchain" bash harness/checks/check_local_toolchain.sh
run_check "local workflow assets" bash harness/checks/check_local_workflow_assets.sh

if [[ "$fail" -eq 0 ]]; then
  echo "$(green '✓ local harness doctor passed')"
else
  echo "$(red '✗ local harness doctor found issues')"
fi

exit "$fail"
