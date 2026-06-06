#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$script_dir/.."
exec bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh "$@"
