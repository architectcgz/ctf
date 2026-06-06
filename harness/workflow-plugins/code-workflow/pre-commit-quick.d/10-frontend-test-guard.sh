#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
bash scripts/check-frontend-test-guard.sh --staged
