#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/.." && pwd)"

exec bash "$HOME/.agents/harness/check-skill-sync-reminder.sh" --cwd "$repo_root" "$@"
