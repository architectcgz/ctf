#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mode="staged"

if [[ "${1:-}" == "--staged" ]]; then
  shift
elif [[ $# -gt 0 ]]; then
  echo "Usage: bash scripts/check-frontend-test-guard.sh --staged" >&2
  exit 2
fi

if [[ "$mode" != "staged" ]]; then
  echo "Only --staged mode is supported." >&2
  exit 2
fi

mapfile -t candidate_files < <(
  git diff --cached --name-only --diff-filter=ACMR -- code/frontend/src \
    | grep -E '^code/frontend/src/.+\.(test|spec)\.(ts|tsx|js|jsx)$' || true
)

if [[ ${#candidate_files[@]} -eq 0 ]]; then
  echo "[frontend-test-guard] no staged frontend test files"
  exit 0
fi

fail=0

matches_for_file() {
  local file="$1"
  git diff --cached --unified=0 -- "$file" \
    | grep -E '^\+[^+]' \
    | grep -En \
      "toContain\(['\"]class=|not\.toContain\(['\"]class=|toMatch\(/.*class=|toContain\(['\"]padding:|toContain\(['\"]gap:"
}

for file in "${candidate_files[@]}"; do
  if grep -q 'ui-test-guard: allow' "$file"; then
    echo "[frontend-test-guard] skip allowlisted file: $file"
    continue
  fi

  if output="$(matches_for_file "$file")" && [[ -n "$output" ]]; then
    echo "[frontend-test-guard] forbidden fine-grained UI test assertions in $file"
    echo "$output"
    fail=1
  fi
done

if [[ "$fail" -ne 0 ]]; then
  cat <<'EOF'

Blocked patterns:
- toContain('class="...') / not.toContain('class="...')
- toMatch(/...class=.../)
- toContain('padding:...') / toContain('gap:...')

Expected direction:
- Prefer behavior, state owner, route/page owner, shared component owner, and architecture boundary tests.
- Use ?raw only for explicit owner/boundary guards, not local markup or style details.

If an exception is truly necessary, add a file comment containing:
ui-test-guard: allow
and explain the owner/boundary reason in the test.
EOF
  exit 1
fi

echo "[frontend-test-guard] staged frontend tests passed"
