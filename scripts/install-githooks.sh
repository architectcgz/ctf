#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

git config core.hooksPath .githooks

chmod +x .githooks/pre-commit
chmod +x scripts/check-reuse-first.sh
chmod +x scripts/check-skill-sync-reminder.sh
chmod +x scripts/sync_openapi_from_contract.py

echo "Installed git hooks to .githooks (core.hooksPath=.githooks)"
