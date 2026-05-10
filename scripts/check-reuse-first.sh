#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

args=("$@")
if [[ "${#args[@]}" -eq 0 ]]; then
  args=(--staged)
fi

python3 harness/checks/check-reuse-decision.py "${args[@]}"
python3 harness/checks/check-similar-pages.py "${args[@]}"
python3 harness/checks/check-duplicate-hooks.py "${args[@]}"
python3 harness/checks/check-api-wrapper-duplication.py "${args[@]}"
