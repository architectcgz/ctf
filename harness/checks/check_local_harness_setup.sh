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

echo "[doctor] git hooks"
hooks_path="$(git config --get core.hooksPath || true)"
check "core.hooksPath is .githooks" test "$hooks_path" = ".githooks"
check ".githooks/pre-commit is executable" test -x ".githooks/pre-commit"
check ".githooks/commit-msg is executable" test -x ".githooks/commit-msg"

echo "[doctor] harness entry scripts"
check "scripts/check-consistency.sh is executable" test -x "scripts/check-consistency.sh"
check "scripts/run-workflow-stage.sh is executable" test -x "scripts/run-workflow-stage.sh"
check "scripts/check-review-governance.sh is executable" test -x "scripts/check-review-governance.sh"
check "scripts/check-review-governance-core.sh is executable" test -x "scripts/check-review-governance-core.sh"
check "scripts/check-task-intake.sh is executable" test -x "scripts/check-task-intake.sh"
check "scripts/start-implementation.sh is executable" test -x "scripts/start-implementation.sh"
check "scripts/check-commit-message.sh is executable" test -x "scripts/check-commit-message.sh"
check "scripts/check-architecture.sh is executable" test -x "scripts/check-architecture.sh"
check "scripts/check-backend-architecture.sh is executable" test -x "scripts/check-backend-architecture.sh"
check "scripts/check-frontend-architecture.sh is executable" test -x "scripts/check-frontend-architecture.sh"
check "scripts/ensure-frontend-tooling.sh is executable" test -x "scripts/ensure-frontend-tooling.sh"
check "scripts/doctor-local-harness.sh is executable" test -x "scripts/doctor-local-harness.sh"

echo "[doctor] harness check bodies"
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
