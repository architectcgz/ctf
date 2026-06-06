run_architecture_checks() {
  echo "[C10] local architecture guardrails are wired"
  check_file "scripts/check-architecture.sh"
  check_file "scripts/check-backend-architecture.sh"
  check_file "scripts/check-frontend-architecture.sh"
  check_file "scripts/check-script-layer.sh"
  check_file "tools/ensure-frontend-tooling.sh"
  check_file "tools/AGENTS.md"
  check_file "scripts/check-task-intake.sh"
  check_file "scripts/check-commit-message.sh"
  check_file "scripts/check-workflow-governance.sh"
  check_file "scripts/check-workflow-governance-core.sh"
  check_file "scripts/check-review-governance.sh"
  check_file "scripts/check-review-governance-core.sh"
  check_file "scripts/check-code-changes.sh"
  check_file "scripts/check-frontend-test-guard.sh"
  check_file "scripts/run-workflow-stage.sh"
  check_file "scripts/check-startup-gate.sh"
  check_file "scripts/check-skill-sync-reminder.sh"
  check_file "scripts/doctor-local-harness.sh"
  check_dir "harness/checks"
  check_file "harness/checks/check_code_change_contracts.sh"
  check_file "harness/checks/check_frontend_test_guard.sh"
  check_file "harness/checks/check_workflow_governance_core.sh"
  check_file "harness/checks/check_local_harness_setup.sh"
  check_file "harness/checks/check_local_toolchain.sh"
  check_file "harness/checks/check_local_workflow_assets.sh"
  check_file "harness/checks/check_script_layer_conventions.py"
  check_file "harness/workflow-plugins/code-workflow/README.md"
  check_file "harness/workflow-plugins/code-workflow/run_workflow_stage.sh"
  check_dir "harness/workflow-plugins/code-workflow/pre-commit-quick.d"
  check_dir "harness/workflow-plugins/code-workflow/completion-full.d"
  check_dir "harness/workflow-plugins/code-workflow/workflow-governance.d"
  check_file "harness/policies/commit-message.json"
  check_file "harness/policies/local-harness-executables.txt"
  check_file "harness/policies/script-layer-manifest.json"
  check_contains ".githooks/pre-commit" 'scripts/run-workflow-stage\.sh pre-commit-quick' "pre-commit runs quick workflow stage"
  check_contains ".githooks/pre-commit" 'scripts/check-skill-sync-reminder\.sh --staged' "pre-commit runs harness reminder separately"
  check_contains ".githooks/commit-msg" 'scripts/check-commit-message\.sh' "commit-msg runs message checks"
  check_contains "scripts/check-workflow-complete.sh" 'scripts/run-workflow-stage\.sh workflow-governance' "workflow complete runs workflow governance stage"
  check_contains "scripts/check-workflow-complete.sh" 'scripts/run-workflow-stage\.sh completion-full' "workflow complete runs completion stage"
  check_contains "scripts/run-workflow-stage.sh" 'harness/workflow-plugins/code-workflow/run_workflow_stage\.sh' "workflow stage wrapper delegates to harness workflow runner"
  check_contains "scripts/check-script-layer.sh" 'check_script_layer_conventions\.py' "script layer wrapper delegates to harness checker"
  check_contains "scripts/check-frontend-architecture.sh" 'tools/ensure-frontend-tooling\.sh' "frontend architecture checks ensure shared tooling"
  check_contains "scripts/check-frontend-test-guard.sh" 'harness/checks/check_frontend_test_guard\.sh' "frontend test guard wrapper delegates to harness checks"
  check_contains "scripts/check-code-changes.sh" 'harness/checks/check_code_change_contracts\.sh' "code change wrapper delegates to harness checks"
  check_contains "scripts/check-workflow-governance-core.sh" 'harness/checks/check_workflow_governance_core\.sh' "workflow governance core wrapper delegates to harness checks"
  check_contains "scripts/check-review-governance-core.sh" 'scripts/check-workflow-governance-core\.sh' "review governance core wrapper stays as compatibility alias"
  check_contains "scripts/doctor-local-harness.sh" 'harness/checks/check_local_harness_setup\.sh' "doctor wrapper runs local harness setup check"
  check_contains "scripts/doctor-local-harness.sh" 'harness/checks/check_local_toolchain\.sh' "doctor wrapper runs local toolchain check"
  check_contains "scripts/doctor-local-harness.sh" 'harness/checks/check_local_workflow_assets\.sh' "doctor wrapper runs local workflow assets check"
  check_contains "harness/checks/check_local_harness_setup.sh" 'scripts/check-agent-entrypoints\.sh' "doctor local harness setup runs entrypoint checks"
  check_contains "scripts/install-githooks.sh" 'harness/policies/local-harness-executables\.txt' "install-githooks reads executable manifest"
  check_contains "harness/checks/check_local_harness_setup.sh" 'harness/policies/local-harness-executables\.txt' "doctor local harness setup reads executable manifest"
  check_contains "harness/checks/check_local_harness_setup.sh" 'scripts/check-script-layer\.sh' "doctor local harness setup runs script layer check"
  check_contains "scripts/install-agent-symlinks.sh" 'scripts/check-agent-entrypoints\.sh' "shared skill installer re-checks entrypoints"
  check_contains "scripts/check-commit-message.sh" 'check_commit_message\.py' "commit message script delegates to shared checker"
  check_contains "scripts/check-commit-message.sh" 'harness/policies/commit-message\.json' "commit message script reads project policy"
  check_contains "scripts/check-skill-sync-reminder.sh" 'remind_skill_sync\.py' "skill sync script delegates to shared reminder"
  check_contains ".githooks/README.md" 'pre-commit-quick' "hook docs mention pre-commit stage"
  check_contains ".githooks/README.md" 'completion-full' "hook docs mention completion stage"
  check_contains ".githooks/README.md" 'workflow-governance' "hook docs mention workflow governance stage"
  check_contains ".githooks/README.md" 'harness 级非阻塞提醒' "hook docs describe harness reminder ownership"
  check_contains ".githooks/README.md" 'commit-msg' "hook docs mention commit-msg"
  check_contains ".githooks/README.md" 'harness/policies/commit-message\.json' "hook docs mention commit message policy"
  check_not_contains ".githooks/README.md" 'harness/bridges/' "hook docs no longer mention generic harness bridges"
  if [[ -e "harness/bridges" ]]; then
    echo "  $(red FAIL) — harness/bridges must not exist; only workflow adapters should live under harness/workflow-plugins/"
    fail=1
  else
    echo "  $(green PASS) — harness/bridges is absent"
  fi
  check_contains "docs/architecture/README.md" 'scripts/check-architecture\.sh --full' "architecture README maps full architecture checks"
  check_contains "docs/architecture/README.md" 'scripts/check-backend-architecture\.sh --full' "architecture README maps backend architecture checks"
  check_contains "docs/architecture/README.md" 'scripts/check-frontend-architecture\.sh --full' "architecture README maps frontend architecture checks"
  check_contains "docs/architecture/README.md" 'harness/workflow-plugins/code-workflow/' "architecture README references workflow plugin registration"
  check_contains "docs/architecture/README.md" 'code/backend/internal/module/architecture_test\.go' "architecture README references backend guardrail test"
  check_contains "docs/architecture/README.md" 'code/backend/internal/app/architecture_rules_test\.go' "architecture README references backend app guardrail test"
  check_contains "docs/architecture/README.md" 'code/frontend/src/__tests__/architectureBoundaries\.test\.ts' "architecture README references frontend guardrail test"
  check_contains "docs/architecture/README.md" 'code/frontend/scripts/check-frontend-growth-guard\.mjs' "architecture README references frontend growth guard"
  check_contains "docs/architecture/README.md" 'code/frontend/scripts/check-vue-deep-guard\.mjs' "architecture README references frontend vue deep guard"
  check_contains "docs/architecture/README.md" 'code/frontend/scripts/frontend-architecture-policy\.json' "architecture README references frontend architecture policy"
  check_contains "docs/architecture/frontend/README.md" 'code/frontend/scripts/check-vue-deep-guard\.mjs' "frontend architecture README references vue deep guard"
  check_contains "docs/architecture/frontend/08-build-deploy.md" 'check:vue-deep' "frontend build doc references vue deep guard script"
  check_contains "code/frontend/package.json" 'check:vue-deep' "frontend package script exposes vue deep guard"
  check_contains "docs/architecture/README.md" 'code/frontend/src/features/contest-awd-admin/model/useAwdOwnerBoundaries\.test\.ts' "architecture README references AWD owner guard"
  check_file "code/frontend/scripts/check-vue-deep-guard.mjs"
  check_file "code/frontend/scripts/vue-deep-allowlist.json"
  check_contains "scripts/check-frontend-architecture.sh" 'check:vue-deep' "frontend architecture script runs vue deep guard"

  echo "[C13] dev compose exposes AWD defense SSH port when enabled"
  dev_config_file="code/backend/configs/config.dev.yaml"
  dev_compose_file="docker/ctf/docker-compose.dev.yml"
  if [[ ! -f "$dev_config_file" ]]; then
    echo "  $(red FAIL) — missing $dev_config_file"
    fail=1
  elif [[ ! -f "$dev_compose_file" ]]; then
    echo "  $(red FAIL) — missing $dev_compose_file"
    fail=1
  elif grep -qE '^[[:space:]]*defense_ssh_enabled:[[:space:]]*true([[:space:]]|$)' "$dev_config_file"; then
    ssh_port="$(
      awk -F: '
        /^[[:space:]]*defense_ssh_port:[[:space:]]*[0-9]+([[:space:]]|$)/ {
          value=$2
          sub(/^[[:space:]]+/, "", value)
          sub(/[[:space:]]+$/, "", value)
          print value
          exit
        }
      ' "$dev_config_file"
    )"
    if [[ -z "$ssh_port" ]]; then
      echo "  $(red FAIL) — unable to resolve defense_ssh_port from $dev_config_file"
      fail=1
    elif grep -qE "\"127\\.0\\.0\\.1:${ssh_port}:${ssh_port}\"|\"0\\.0\\.0\\.0:${ssh_port}:${ssh_port}\"|\"${ssh_port}:${ssh_port}\"" "$dev_compose_file"; then
      echo "  $(green PASS) — dev compose exposes AWD defense SSH port $ssh_port"
    else
      echo "  $(red FAIL) — dev compose must expose AWD defense SSH port $ssh_port when defense_ssh_enabled is true"
      fail=1
    fi
  else
    echo "  $(green PASS) — dev defense SSH gateway disabled"
  fi

  echo "[C14] AWD runtime prioritizes /workspace/src over image fallback"
  runtime_app_files="$(
    find challenges/awd -path '*/docker/runtime/app.py' | sort
  )"
  if [[ -z "${runtime_app_files// }" ]]; then
    echo "  $(green PASS) — no AWD runtime app.py files found"
  else
    while IFS= read -r file; do
      [[ -z "$file" ]] && continue
      if grep -qF 'if str(WORKSPACE_SRC) not in sys.path' "$file"; then
        echo "  $(red FAIL) — $file still uses conditional WORKSPACE_SRC insertion"
        fail=1
        continue
      fi
      if grep -qF 'workspace_src = str(WORKSPACE_SRC)' "$file" \
        && grep -qF 'sys.path = [workspace_src] + [path for path in sys.path if path != workspace_src]' "$file"; then
        echo "  $(green PASS) — $file pins /workspace/src ahead of image fallback"
      else
        echo "  $(red FAIL) — $file must deduplicate and pin /workspace/src at sys.path[0]"
        fail=1
      fi
    done <<< "$runtime_app_files"
  fi
}
