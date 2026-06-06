#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

usage() {
  cat <<'EOF' >&2
Usage:
  bash scripts/check-frontend-test-guard.sh
  bash scripts/check-frontend-test-guard.sh --working-tree
  bash scripts/check-frontend-test-guard.sh --staged
  bash scripts/check-frontend-test-guard.sh --files <path...>
EOF
}

mode="working-tree"
declare -a requested_paths=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --working-tree)
      mode="working-tree"
      shift
      ;;
    --staged)
      mode="staged"
      shift
      ;;
    --files)
      mode="files"
      shift
      if [[ $# -eq 0 ]]; then
        usage
        exit 2
      fi
      while [[ $# -gt 0 ]]; do
        requested_paths+=("$1")
        shift
      done
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

readonly TEST_FILE_PATTERN='^code/frontend/src/.+\.(test|spec)\.(ts|tsx|js|jsx)$'
readonly BLOCKED_PATTERN="toContain\(['\"]class=|not\.toContain\(['\"]class=|toMatch\(/.*class=|toContain\(['\"]padding:|toContain\(['\"]gap:"

is_frontend_test_file() {
  local file="$1"
  [[ "$file" =~ $TEST_FILE_PATTERN ]]
}

normalize_path() {
  local path="$1"
  if [[ "$path" == "$PWD/"* ]]; then
    printf '%s\n' "${path#$PWD/}"
    return
  fi

  if [[ "$path" == "$PWD" ]]; then
    printf '.\n'
    return
  fi

  printf '%s\n' "$path"
}

collect_staged_files() {
  git diff --cached --name-only --diff-filter=ACMR -- code/frontend/src \
    | grep -E "$TEST_FILE_PATTERN" || true
}

collect_working_tree_files() {
  {
    git diff --name-only --diff-filter=ACMR HEAD -- code/frontend/src
    git ls-files --others --exclude-standard -- code/frontend/src
  } | grep -E "$TEST_FILE_PATTERN" | sort -u || true
}

collect_explicit_files() {
  local path=""
  for path in "${requested_paths[@]}"; do
    path="$(normalize_path "$path")"

    if [[ -f "$path" ]]; then
      if is_frontend_test_file "$path"; then
        printf '%s\n' "$path"
      else
        echo "[frontend-test-guard] ignore non-frontend-test file: $path" >&2
      fi
      continue
    fi

    if [[ -d "$path" ]]; then
      rg --files "$path" | grep -E "$TEST_FILE_PATTERN" || true
      continue
    fi

    echo "[frontend-test-guard] missing path: $path" >&2
  done
}

grep_matches() {
  grep -En "$BLOCKED_PATTERN" || true
}

matches_for_staged_file() {
  local file="$1"
  git diff --cached --unified=0 -- "$file" \
    | grep -E '^\+[^+]' \
    | grep_matches || true
}

matches_for_working_tree_file() {
  local file="$1"

  if git ls-files --error-unmatch -- "$file" >/dev/null 2>&1; then
    git diff --unified=0 HEAD -- "$file" \
      | grep -E '^\+[^+]' \
      | grep_matches || true
    return
  fi

  grep -En "$BLOCKED_PATTERN" "$file" || true
}

matches_for_explicit_file() {
  local file="$1"
  grep -En "$BLOCKED_PATTERN" "$file" || true
}

declare -a candidate_files=()
case "$mode" in
  staged)
    mapfile -t candidate_files < <(collect_staged_files)
    ;;
  working-tree)
    mapfile -t candidate_files < <(collect_working_tree_files)
    ;;
  files)
    mapfile -t candidate_files < <(collect_explicit_files | sort -u)
    ;;
esac

if [[ ${#candidate_files[@]} -eq 0 ]]; then
  echo "[frontend-test-guard] no frontend test files for mode: $mode"
  exit 0
fi

fail=0

for file in "${candidate_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "[frontend-test-guard] skip missing file: $file"
    continue
  fi

  if grep -q 'ui-test-guard: allow' "$file"; then
    echo "[frontend-test-guard] skip allowlisted file: $file"
    continue
  fi

  output=""
  case "$mode" in
    staged)
      output="$(matches_for_staged_file "$file")"
      ;;
    working-tree)
      output="$(matches_for_working_tree_file "$file")"
      ;;
    files)
      output="$(matches_for_explicit_file "$file")"
      ;;
  esac

  if [[ -n "$output" ]]; then
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

echo "[frontend-test-guard] frontend tests passed for mode: $mode"
