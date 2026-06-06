#!/usr/bin/env bash
set -euo pipefail

matches_any() {
  local pattern="$1"
  local input="$2"
  if command -v rg >/dev/null 2>&1; then
    printf '%s\n' "$input" | rg -q "$pattern"
  else
    printf '%s\n' "$input" | grep -Eq "$pattern"
  fi
}

changed="${WORKFLOW_CHANGED_FILES:-}"
pattern='(^docs/architecture/|^scripts/check-architecture\.sh$|^scripts/check-frontend-architecture\.sh$|^code/frontend/package\.json$|^code/frontend/scripts/|^code/frontend/src/|^code/frontend/vite\.config\.[^/]+$)'

cd "$(git rev-parse --show-toplevel)"

if [[ -n "$changed" ]] && matches_any "$pattern" "$changed"; then
  bash scripts/check-frontend-architecture.sh --full
else
  echo "no frontend architecture-sensitive changes detected"
fi
