#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
EXECUTABLE_LIST="harness/policies/local-harness-executables.txt"

cd "$ROOT_DIR"

git config core.hooksPath .githooks

if [[ ! -f "$EXECUTABLE_LIST" ]]; then
  echo "missing executable manifest: $EXECUTABLE_LIST" >&2
  exit 1
fi

while IFS= read -r path; do
  [[ -z "$path" ]] && continue
  if [[ ! -e "$path" ]]; then
    echo "missing executable entrypoint: $path" >&2
    exit 1
  fi
  chmod +x "$path"
done < "$EXECUTABLE_LIST"

echo "Installed git hooks to .githooks (core.hooksPath=.githooks)"
