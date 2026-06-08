#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
EXECUTABLE_LIST="harness/policies/local-harness-executables.txt"
GLOBAL_ENTRYPOINT_CHECK="$HOME/.agents/harness/check-project-agent-entrypoints.sh"
GLOBAL_SHARED_SKILLS_CHECK="$HOME/.agents/harness/check-project-shared-skills.sh"
GLOBAL_SKILL_INSTALLER="$HOME/.agents/harness/install-project-skills.sh"
GLOBAL_SKILL_UNINSTALLER="$HOME/.agents/harness/uninstall-project-skills.sh"
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
check "harness/checks/check_frontend_test_guard.sh exists" test -f "harness/checks/check_frontend_test_guard.sh"
check "harness/checks/check_test_architecture_guard.py exists" test -f "harness/checks/check_test_architecture_guard.py"
check "harness/checks/check_workflow_governance_core.sh exists" test -f "harness/checks/check_workflow_governance_core.sh"
check "harness/checks/check_local_harness_setup.sh exists" test -f "harness/checks/check_local_harness_setup.sh"
check "harness/checks/check_local_toolchain.sh exists" test -f "harness/checks/check_local_toolchain.sh"
check "harness/checks/check_local_workflow_assets.sh exists" test -f "harness/checks/check_local_workflow_assets.sh"
check "harness/checks/check_script_layer_conventions.py exists" test -f "harness/checks/check_script_layer_conventions.py"
check "harness/workflow-plugins/code-workflow/run_workflow_stage.sh exists" test -f "harness/workflow-plugins/code-workflow/run_workflow_stage.sh"
check "harness/workflow-plugins/code-workflow/archive_task_artifacts.sh exists" test -f "harness/workflow-plugins/code-workflow/archive_task_artifacts.sh"

echo "[doctor] local agent and skill wiring"
check "agent entrypoints stay aligned" bash scripts/check-agent-entrypoints.sh
check "shared skills stay aligned" bash scripts/check-shared-skills.sh

echo "[doctor] shared harness skill tooling"
check "$GLOBAL_ENTRYPOINT_CHECK is executable" test -x "$GLOBAL_ENTRYPOINT_CHECK"
check "$GLOBAL_SHARED_SKILLS_CHECK is executable" test -x "$GLOBAL_SHARED_SKILLS_CHECK"
check "$GLOBAL_SKILL_INSTALLER is executable" test -x "$GLOBAL_SKILL_INSTALLER"
check "$GLOBAL_SKILL_UNINSTALLER is executable" test -x "$GLOBAL_SKILL_UNINSTALLER"

echo "[doctor] workflow wiring"
check "harness/policies/commit-message.json exists" test -f "harness/policies/commit-message.json"
check "harness/policies/script-layer-manifest.json exists" test -f "harness/policies/script-layer-manifest.json"
check "workflow plugin root exists" test -d "harness/workflow-plugins/code-workflow"
check "pre-commit stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/pre-commit-quick.d"
check "completion stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/completion-full.d"
check "workflow governance stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/workflow-governance.d"
check "script layer check passes" bash scripts/check-script-layer.sh

exit "$fail"
