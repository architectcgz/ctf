#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

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
