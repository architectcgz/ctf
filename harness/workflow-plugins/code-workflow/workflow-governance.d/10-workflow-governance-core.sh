#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
bash scripts/check-workflow-governance-core.sh
