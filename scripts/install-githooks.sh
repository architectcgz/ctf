#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

git config core.hooksPath .githooks

chmod +x .githooks/pre-commit
chmod +x .githooks/commit-msg
chmod +x scripts/check-consistency.sh
chmod +x scripts/check-review-governance.sh
chmod +x scripts/check-frontend-test-guard.sh
chmod +x scripts/check-architecture.sh
chmod +x scripts/check-commit-message.sh
chmod +x scripts/check-skill-sync-reminder.sh
chmod +x scripts/ensure-frontend-tooling.sh
chmod +x scripts/sync_openapi_from_contract.py

echo "Installed git hooks to .githooks (core.hooksPath=.githooks)"
