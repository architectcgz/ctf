#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

fail=0

source "scripts/lib/check-consistency/common.sh"
source "scripts/lib/check-consistency/navigation.sh"
source "scripts/lib/check-consistency/architecture.sh"
source "scripts/lib/check-consistency/workflow.sh"

run_navigation_checks
run_architecture_checks
run_workflow_checks

if [[ "$fail" -eq 0 ]]; then
  echo "$(green '✓ review governance checks passed')"
else
  echo "$(red '✗ review governance checks failed')"
fi

exit "$fail"
