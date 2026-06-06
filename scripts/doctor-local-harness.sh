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
check ".githooks/commit-msg is executable" test -x ".githooks/commit-msg"

echo "[doctor] harness scripts"
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
check "harness/policies/commit-message.json exists" test -f "harness/policies/commit-message.json"
check "workflow plugin root exists" test -d "harness/workflow-plugins/code-workflow"
check "pre-commit stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/pre-commit-quick.d"
check "completion stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/completion-full.d"
check "review stage plugin dir exists" test -d "harness/workflow-plugins/code-workflow/review-governance.d"

echo "[doctor] local tools"
check "git is available" command -v git
check "python3 is available" command -v python3
check "go is available" command -v go
check "node is available" command -v node
check "npm is available" command -v npm

echo "[doctor] local workflow assets"
check ".harness/session-gates directory is available or can be created" test -d ".harness/session-gates" -o ! -e ".harness/session-gates"
check "legacy .harness/reuse-decision.md is absent" test ! -f ".harness/reuse-decision.md"
check ".gitignore ignores .harness/reuse-index/" grep -qx '/.harness/reuse-index/' ".gitignore"
check ".gitignore ignores .harness/session-gates/" grep -qx '/.harness/session-gates/' ".gitignore"

if bash scripts/ensure-frontend-tooling.sh --quiet >/dev/null 2>&1; then
  echo "  $(green PASS) — frontend tooling is available"
else
  echo "  $(red FAIL) — frontend tooling is unavailable; run npm install in code/frontend or ensure main worktree dependencies exist"
  fail=1
fi

if [[ "$fail" -eq 0 ]]; then
  echo "$(green '✓ local harness doctor passed')"
else
  echo "$(red '✗ local harness doctor found issues')"
fi

exit "$fail"
