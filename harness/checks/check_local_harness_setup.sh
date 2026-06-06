#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
EXECUTABLE_LIST="harness/policies/local-harness-executables.txt"
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
check "$EXECUTABLE_LIST exists" test -f "$EXECUTABLE_LIST"

echo "[doctor] local harness entrypoints"
while IFS= read -r path; do
  [[ -z "$path" ]] && continue
  check "$path is executable" test -x "$path"
done < "$EXECUTABLE_LIST"

echo "[doctor] harness check bodies"
check "harness/checks/check_code_change_contracts.sh exists" test -f "harness/checks/check_code_change_contracts.sh"
check "harness/checks/check_review_governance_core.sh exists" test -f "harness/checks/check_review_governance_core.sh"
check "harness/checks/check_local_harness_setup.sh exists" test -f "harness/checks/check_local_harness_setup.sh"
check "harness/checks/check_local_toolchain.sh exists" test -f "harness/checks/check_local_toolchain.sh"
check "harness/checks/check_local_workflow_assets.sh exists" test -f "harness/checks/check_local_workflow_assets.sh"

echo "[doctor] workflow wiring"
check "harness/policies/commit-message.json exists" test -f "harness/policies/commit-message.json"
check "workflow plugin root exists" test -d "harness/workflow-plugins/code-workflow"
check "pre-commit stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/pre-commit-quick.d"
check "completion stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/completion-full.d"
check "review stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/review-governance.d"

exit "$fail"
