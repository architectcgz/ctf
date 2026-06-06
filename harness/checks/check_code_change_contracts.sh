#!/usr/bin/env bash
set -euo pipefail

fail=0

red() { printf '\033[31m%s\033[0m' "$1"; }
green() { printf '\033[32m%s\033[0m' "$1"; }

changed_files() {
  local staged
  staged="$(git diff --cached --name-only)"
  if [[ -n "$staged" ]]; then
    printf '%s\n' "$staged" | sort -u
    return
  fi

  git diff --name-only | sort -u
}

matches_any() {
  local pattern="$1"
  local input="$2"
  if command -v rg >/dev/null 2>&1; then
    printf '%s\n' "$input" | rg -q "$pattern"
  else
    printf '%s\n' "$input" | grep -Eq "$pattern"
  fi
}

check_api_contract_changes() {
  local changed="$1"
  local runtime_changed
  runtime_changed="$(printf '%s\n' "$changed" | grep -Ev '^code/frontend/src/api/__tests__/' || true)"
  local api_surface_pattern='(^code/backend/internal/app/router\.go$|^code/backend/internal/dto/|^code/backend/internal/handler/|^code/backend/internal/module/.+/api/|^code/backend/internal/module/.+/handler\.go$|^code/backend/pkg/response/|^code/backend/pkg/errcode/|^code/frontend/src/api/.+\.(ts|vue)$|^code/frontend/src/api/contracts\.ts$)'
  local contract_pattern='(^docs/contracts/api-contract-v1\.md$|^docs/contracts/openapi-v1\.yaml$|^docs/contracts/openapi-v1/|^docs/architecture/backend/04-api-design\.md$)'

  if ! matches_any "$api_surface_pattern" "$runtime_changed"; then
    echo "  $(green PASS) — API surface unchanged"
    return
  fi

  if matches_any "$contract_pattern" "$changed"; then
    echo "  $(green PASS) — API surface change includes contract documentation"
    return
  fi

  echo "  $(red FAIL) — API surface changed without updating API contract docs"
  echo "    Update at least one of:"
  echo "    - docs/contracts/api-contract-v1.md"
  echo "    - docs/contracts/openapi-v1.yaml"
  echo "    - docs/contracts/openapi-v1/"
  echo "    - docs/architecture/backend/04-api-design.md"
  fail=1
}

main() {
  cd "$(git rev-parse --show-toplevel)"

  echo "[code-changes] code change checks"
  local changed
  changed="$(changed_files)"

  if [[ -z "$changed" ]]; then
    echo "  $(green PASS) — no changed files"
    return 0
  fi

  check_api_contract_changes "$changed"

  if [[ "$fail" -eq 0 ]]; then
    echo "$(green '✓ code change checks passed')"
  else
    echo "$(red '✗ code change checks failed')"
  fi

  return "$fail"
}

main "$@"
