#!/usr/bin/env bash
set -euo pipefail

strict=0
quiet_if_empty=0

for arg in "$@"; do
  case "$arg" in
    --strict)
      strict=1
      ;;
    --quiet-if-empty)
      quiet_if_empty=1
      ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 2
      ;;
  esac
done

todo_dir=""
for candidate in docs/todos docs/todo todo; do
  if [[ -d "$candidate" ]]; then
    todo_dir="$candidate"
    break
  fi
done

if [[ -z "$todo_dir" ]]; then
  if [[ "$quiet_if_empty" -eq 0 ]]; then
    echo "[todo] no todo directory found"
  fi
  exit 0
fi

matches="$(
  rg -n '^- \[ \]' "$todo_dir" \
    --glob '*.md' \
    --glob '!README.md' \
    --glob '!active.md' \
    || true
)"

todo_count="$(printf '%s\n' "$matches" | sed '/^$/d' | wc -l | tr -d ' ')"
if [[ "$todo_count" -eq 0 ]]; then
  if [[ "$quiet_if_empty" -eq 0 ]]; then
    echo "[todo] no open todo items in $todo_dir"
  fi
  exit 0
fi

file_count="$(printf '%s\n' "$matches" | sed '/^$/d' | cut -d: -f1 | sort -u | wc -l | tr -d ' ')"

echo "[todo] reminder: $todo_count open items across $file_count files in $todo_dir"
shown=0
while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  file="${line%%:*}"
  rest="${line#*:}"
  lineno="${rest%%:*}"
  text="${rest#*:}"
  text="${text#- [ ] }"
  echo "  - ${file}:${lineno} ${text}"
  shown=$((shown + 1))
  if [[ "$shown" -ge 20 ]]; then
    break
  fi
done <<< "$matches"

if [[ "$todo_count" -gt "$shown" ]]; then
  echo "  - ... and $((todo_count - shown)) more open items"
fi

if [[ "$strict" -eq 1 ]]; then
  exit 2
fi
