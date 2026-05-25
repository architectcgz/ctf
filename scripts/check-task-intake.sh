#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

reuse_document=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --reuse-decision)
      if [[ $# -lt 2 ]]; then
        echo "FAIL: --reuse-decision requires a task slug or document path" >&2
        exit 1
      fi
      reuse_document="$2"
      shift 2
      ;;
    *)
      echo "FAIL: unknown argument: $1" >&2
      echo "usage: bash scripts/check-task-intake.sh [--reuse-decision <task-slug-or-path>]" >&2
      exit 1
      ;;
  esac
done

if [[ -x "scripts/check-open-todos.sh" ]]; then
  bash scripts/check-open-todos.sh --quiet-if-empty
fi

if [[ -n "$reuse_document" ]]; then
  python3 harness/checks/check-reuse-startup.py --document "$reuse_document"
else
  echo "PASS: task intake reminder completed"
  echo "- no startup reuse-decision gate requested"
fi
