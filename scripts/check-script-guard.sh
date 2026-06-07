#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cwd="$(cd "$script_dir/.." && pwd)"
python3 ~/.agents/harness/checks/check_script_guard.py --cwd "$cwd" "$@"
