run_workflow_checks() {
  echo "[C11] startup workflow scaffold is wired"
  check_dir ".harness"
  check_file "harness/templates/implementation-plan-skeleton.md"
  check_file "harness/checks/change_detection.py"
  check_file "harness/checks/common.py"
  check_file "harness/checks/check_startup_gate.py"
  check_file "harness/workflow-plugins/code-workflow/archive_task_artifacts.sh"
  check_file "scripts/check-task-intake.sh"
  check_file "scripts/start-implementation.sh"
  check_file "scripts/check-startup-gate.sh"
  check_contains "AGENTS.md" 'scripts/start-implementation\.sh' "AGENTS references implementation startup gate"
  check_contains "AGENTS.md" 'harness/workflow-plugins/code-workflow/archive_task_artifacts\.sh' "AGENTS references workflow archive entry"
  check_contains "AGENTS.md" '\.harness/session-gates/' "AGENTS references local startup gate directory"
  check_contains "AGENTS.md" '\.harness/reuse-index/' "AGENTS references local private reuse index"
  check_contains "AGENTS.md" '\.harness/reuse-decisions/' "AGENTS still explains optional supplemental reuse notes"
  if grep -qx '/.harness/reuse-index/' ".gitignore"; then
    echo "  $(green PASS) — .gitignore reserves local private reuse index"
  else
    echo "  $(red FAIL) — .gitignore must ignore /.harness/reuse-index/"
    fail=1
  fi
  if grep -qx '/.harness/session-gates/' ".gitignore"; then
    echo "  $(green PASS) — .gitignore reserves local startup gate directory"
  else
    echo "  $(red FAIL) — .gitignore must ignore /.harness/session-gates/"
    fail=1
  fi
  if [[ -d ".harness/reuse-index" ]]; then
    echo "  $(green PASS) — local private reuse index exists"
  else
    echo "  $(green PASS) — local private reuse index is optional and currently absent"
  fi
  if [[ -f ".harness/reuse-decision.md" ]]; then
    echo "  $(red FAIL) — legacy .harness/reuse-decision.md is forbidden"
    fail=1
  else
    echo "  $(green PASS) — no legacy single-file reuse decision present"
  fi
  if [[ -f "scripts/archive-task-artifacts.sh" ]]; then
    echo "  $(red FAIL) — legacy scripts/archive-task-artifacts.sh must be removed"
    fail=1
  else
    echo "  $(green PASS) — no legacy archive wrapper remains in scripts/"
  fi
  check_contains "AGENTS.md" 'scripts/start-implementation\.sh.*startup gate|startup gate.*scripts/start-implementation\.sh' "AGENTS marks startup gate as authoritative"

  echo "[C12] changed feedback records declare sedimentation status"
  feedback_changed="$(
    {
      git diff --name-only --diff-filter=ACMR HEAD -- 'feedback/*.md' 2>/dev/null || true
      git ls-files --others --exclude-standard -- 'feedback/*.md' 2>/dev/null || true
    } | sort -u
  )"
  if [[ -z "${feedback_changed// }" ]]; then
    echo "  $(green PASS) — no changed feedback records"
  else
    while IFS= read -r file; do
      [[ -z "$file" || "$file" == "feedback/AGENTS.md" || "$file" == "feedback/improvements-index.md" ]] && continue
      check_contains "$file" '^## 沉淀状态$' "$file declares sedimentation status"
    done <<< "$feedback_changed"
  fi
}
