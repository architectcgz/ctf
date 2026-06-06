#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cwd="$(cd "$script_dir/.." && pwd)"
python3 ~/.agents/harness/skill-sync/remind_skill_sync.py --cwd "$cwd" "$@"
