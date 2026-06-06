#!/usr/bin/env bash
set -euo pipefail

quiet=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --quiet)
      quiet=1
      shift
      ;;
    -h|--help)
      cat <<'EOF'
Usage:
  bash tools/ensure-frontend-tooling.sh [--quiet]

Description:
  Ensure code/frontend/node_modules is available in the current worktree.
  If the current worktree is missing frontend dependencies, the script will
  reuse the main worktree's node_modules via a symlink when possible.
EOF
      exit 0
      ;;
    *)
      echo "FAIL: unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/code/frontend"
NODE_MODULES_DIR="$FRONTEND_DIR/node_modules"

log() {
  if [[ "$quiet" -eq 0 ]]; then
    echo "$@"
  fi
}

if [[ -d "$NODE_MODULES_DIR" ]]; then
  log "[frontend-tooling] node_modules available"
  exit 0
fi

current_root="$(git -C "$ROOT_DIR" rev-parse --show-toplevel)"
shared_node_modules=""

while IFS= read -r worktree_path; do
  [[ -z "$worktree_path" ]] && continue
  if [[ "$worktree_path" == "$current_root" ]]; then
    continue
  fi
  candidate="$worktree_path/code/frontend/node_modules"
  if [[ -d "$candidate" ]]; then
    shared_node_modules="$candidate"
    break
  fi
done < <(git -C "$ROOT_DIR" worktree list --porcelain | awk '/^worktree / {print substr($0, 10)}')

if [[ -n "$shared_node_modules" ]]; then
  ln -s "$shared_node_modules" "$NODE_MODULES_DIR"
  log "[frontend-tooling] linked shared node_modules from $shared_node_modules"
  exit 0
fi

echo "FAIL: code/frontend/node_modules is missing in this worktree" >&2
echo "Run 'cd code/frontend && npm install' in the main worktree or this worktree before continuing." >&2
exit 1
