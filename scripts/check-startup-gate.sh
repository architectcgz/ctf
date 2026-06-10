#!/usr/bin/env bash
# Managed by code-workflow package (version: 2026-06-10.1)
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$script_dir/.."

python3 harness/checks/check_startup_gate.py "$@"
