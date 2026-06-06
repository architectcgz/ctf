#!/usr/bin/env bash
# Managed by code-workflow package (version: 2026-06-06.2)
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$script_dir/.."

if [[ $# -gt 0 ]]; then
  echo "FAIL: check-task-intake.sh no longer accepts task-scoped reuse arguments" >&2
  echo "Use: bash scripts/start-implementation.sh <topic-or-slug>" >&2
  exit 1
fi
if [[ -x "scripts/check-open-todos.sh" ]]; then
  bash scripts/check-open-todos.sh --quiet-if-empty
fi

echo "PASS: task intake reminder completed"
echo "- reviewed open todos and local intake reminders"
echo "- non-trivial or protected implementation should start with: bash scripts/start-implementation.sh <topic-or-slug>"
echo "- before finalizing the plan, run the intake analysis gate: relevant superpowers analysis pass first, then grill-with-docs"
