run_navigation_checks() {
  echo "[C1] strict harness directories exist"
  for dir in concepts thinking practice feedback works references; do
    check_dir "$dir"
    check_file "$dir/AGENTS.md"
  done

  echo "[C2] root navigation references strict harness"
  check_contains "AGENTS.md" 'concepts/' "AGENTS references concepts"
  check_contains "AGENTS.md" 'thinking/' "AGENTS references thinking"
  check_contains "AGENTS.md" 'practice/' "AGENTS references practice"
  check_contains "AGENTS.md" 'feedback/' "AGENTS references feedback"
  check_contains "AGENTS.md" 'works/' "AGENTS references works"
  check_contains "AGENTS.md" 'harness/bridges/' "AGENTS references harness bridges"
  check_contains "AGENTS.md" 'harness/prompts/' "AGENTS references harness prompts"
  check_contains "AGENTS.md" 'references/' "AGENTS references references"

  echo "[C2a] project CLAUDE entrypoint stays aligned"
  if [[ -L "CLAUDE.md" ]]; then
    if [[ "$(readlink -f CLAUDE.md)" == "$(readlink -f AGENTS.md)" ]]; then
      echo "  $(green PASS) — CLAUDE.md points to AGENTS.md"
    else
      echo "  $(red FAIL) — CLAUDE.md does not resolve to AGENTS.md"
      fail=1
    fi
  else
    echo "  $(red FAIL) — CLAUDE.md must be a symlink to AGENTS.md"
    fail=1
  fi

  echo "[C2b] shared skill bridge stays aligned"
  check_file "scripts/check-agent-entrypoints.sh"
  check_file "scripts/check-shared-skills.sh"
  check_dir "harness/checks"
  check_file "harness/checks/check_agent_entrypoints.sh"
  check_file "harness/checks/check_shared_skills.sh"
  check_file "scripts/install-agent-symlinks.sh"
  check_file "scripts/uninstall-agent-symlinks.sh"
  check_dir "harness/bridges"
  check_file "harness/bridges/README.md"
  check_file "harness/bridges/install-agent-symlinks.sh"
  check_file "harness/bridges/uninstall-agent-symlinks.sh"
  check_contains "AGENTS.md" '\.agents/skills/' "AGENTS references shared skill source"
  check_contains "AGENTS.md" 'scripts/install-agent-symlinks\.sh' "AGENTS references shared skill installer"
  check_contains "scripts/check-agent-entrypoints.sh" 'harness/checks/check_agent_entrypoints\.sh' "check-agent-entrypoints wrapper delegates to harness check"
  check_contains "scripts/check-shared-skills.sh" 'harness/checks/check_shared_skills\.sh' "check-shared-skills wrapper delegates to harness check"
  check_contains "scripts/install-agent-symlinks.sh" 'harness/bridges/install-agent-symlinks\.sh' "install-agent-symlinks wrapper delegates to harness bridge"
  check_contains "scripts/uninstall-agent-symlinks.sh" 'harness/bridges/uninstall-agent-symlinks\.sh' "uninstall-agent-symlinks wrapper delegates to harness bridge"
  if [[ -x "scripts/check-agent-entrypoints.sh" ]]; then
    bash scripts/check-agent-entrypoints.sh
  else
    echo "  $(red FAIL) — scripts/check-agent-entrypoints.sh is not executable"
    fail=1
  fi
  if [[ -x "scripts/check-shared-skills.sh" ]]; then
    bash scripts/check-shared-skills.sh
  else
    echo "  $(red FAIL) — scripts/check-shared-skills.sh is not executable"
    fail=1
  fi

  echo "[C3] articles.md numbering is contiguous 1..N"
  nums=$(grep -nE '^### [0-9]+\.' references/articles.md | sed -E 's/^[0-9]+:### ([0-9]+)\..*/\1/' || true)
  count=$(echo "$nums" | sed '/^$/d' | wc -l | tr -d ' ')
  if [[ "$count" -eq 0 ]]; then
    echo "  $(red FAIL) — references/articles.md has no numbered entries"
    fail=1
  else
    sorted=$(echo "$nums" | sort -n)
    expected=$(seq 1 "$count")
    if [[ "$sorted" = "$expected" ]]; then
      echo "  $(green PASS) — $count contiguous entries"
    else
      echo "  $(red FAIL) — article numbering is not contiguous"
      fail=1
    fi
  fi

  echo "[C4] article count claim matches numbered entries"
  claim=$(grep -oE '权威计数：[0-9]+ 篇' references/articles.md | head -1 | grep -oE '[0-9]+' || true)
  if [[ -z "$claim" || "$claim" != "$count" ]]; then
    echo "  $(red FAIL) — references/articles.md claims ${claim:-none}, actual $count"
    fail=1
  else
    echo "  $(green PASS) — count claim $claim"
  fi

  echo "[C5] hook runs strict consistency check"
  if [[ -f ".githooks/pre-commit" ]]; then
    check_contains ".githooks/pre-commit" 'scripts/run-workflow-stage\.sh pre-commit-quick' "pre-commit runs workflow stage runner"
    check_contains ".githooks/pre-commit" 'scripts/check-skill-sync-reminder\.sh --staged' "pre-commit runs harness skill sync reminder"
    check_not_contains ".githooks/pre-commit" 'scripts/check-reuse-first\.sh --staged' "pre-commit no longer runs project-local reuse-first"
  else
    echo "  $(red FAIL) — missing .githooks/pre-commit"
    fail=1
  fi
  check_file ".githooks/commit-msg"
  check_file "scripts/check-commit-message.sh"
  check_file "scripts/check-review-governance.sh"
  check_file "scripts/check-review-governance-core.sh"
  check_file "scripts/run-workflow-stage.sh"
  check_file "harness/policies/commit-message.json"
  check_file "scripts/check-frontend-test-guard.sh"
  check_file "scripts/check-startup-gate.sh"
  check_contains ".githooks/commit-msg" 'scripts/check-commit-message\.sh' "commit-msg runs scripts/check-commit-message.sh"
  check_contains "scripts/check-consistency.sh" 'scripts/check-review-governance\.sh' "check-consistency delegates to review governance"
  check_contains "scripts/check-review-governance.sh" 'scripts/run-workflow-stage\.sh review-governance' "review governance wrapper uses stage runner"
  check_contains ".githooks/README.md" 'scripts/check-commit-message\.sh' "hook docs mention commit message checks"
  check_contains ".githooks/README.md" 'harness/policies/commit-message\.json' "hook docs mention commit message policy"
  check_contains ".githooks/README.md" 'scripts/run-workflow-stage\.sh pre-commit-quick' "hook docs mention workflow stage runner"
  check_contains ".githooks/README.md" 'scripts/check-skill-sync-reminder\.sh --staged' "hook docs mention harness skill sync reminder"
  check_contains ".githooks/README.md" 'bash scripts/check-frontend-test-guard\.sh' "hook docs mention manual frontend test guard workflow"
  check_contains ".githooks/README.md" 'check-frontend-test-guard\.sh --files' "hook docs mention explicit frontend test guard file mode"
  check_not_contains ".githooks/README.md" 'scripts/check-reuse-first\.sh --staged' "hook docs no longer mention project-local reuse-first gate"

  echo "[C6] migrated accumulation indexes exist"
  check_file "feedback/improvements-index.md"
  check_file "practice/superpowers-plan-index.md"
  check_file "practice/planning-archive-index.md"
  check_file "harness/prompts/AGENTS.md"
  check_file "harness/prompts/architecture-diagram-generation.md"
  check_file "references/ctf-instance-lifecycle-research.md"
  check_file "works/harness-migration-map.md"

  echo "[C7] migrated indexes are discoverable from directory AGENTS"
  check_contains "feedback/AGENTS.md" 'improvements-index\.md' "feedback AGENTS references migrated improvements"
  check_contains "practice/AGENTS.md" 'superpowers-plan-index\.md' "practice AGENTS references superpowers index"
  check_contains "practice/AGENTS.md" 'planning-archive-index\.md' "practice AGENTS references planning archive"
  check_contains "harness/prompts/AGENTS.md" 'architecture-diagram-generation\.md' "harness prompts AGENTS references architecture diagram prompt"
  check_not_contains "harness/prompts/AGENTS.md" 'coding-agent-system-prompt\.md' "harness prompts AGENTS no longer expose project-local reuse-first prompt"
  check_contains "references/AGENTS.md" 'ctf-instance-lifecycle-research\.md' "references AGENTS references lifecycle research"
  check_contains "works/AGENTS.md" 'harness-migration-map\.md' "works AGENTS references migration map"

  echo "[C8] AGENTS captures file placement rules"
  check_file "docs/文档规范.md"
  check_file "docs/contracts/README.md"
  check_file "docs/contracts/openapi-v1/index.yaml"
  check_file "scripts/check-task-intake.sh"
  check_file "scripts/start-implementation.sh"
  check_file "scripts/check-open-todos.sh"
  check_file "scripts/check-docs-consistency.py"
  check_file "harness/checks/check_open_todos.sh"
  check_file "harness/checks/check_docs_consistency.py"
  check_file "scripts/sync_openapi_from_contract.py"
  check_contains "AGENTS.md" 'docs/文档规范\.md' "AGENTS references documentation guide"
  check_contains "docs/README.md" 'docs/文档规范\.md' "docs README references documentation guide"
  check_contains "docs/README.md" 'docs/contracts/README\.md' "docs README references contracts guide"
  check_contains "docs/contracts/README.md" 'openapi-v1/' "contracts README references split OpenAPI source"
  check_contains "docs/contracts/api-contract-v1.md" 'openapi-v1/' "API contract references split OpenAPI source"
  check_contains "docs/文档规范.md" '文档修改前置读取协议' "documentation guide defines pre-edit reading protocol"
  check_contains "docs/文档规范.md" '新增路径登记协议' "documentation guide defines new path registration protocol"
  check_contains "AGENTS.md" '文档修改前置读取协议' "AGENTS references document pre-edit protocol"
  check_contains "AGENTS.md" '新增路径登记协议' "AGENTS references new path registration protocol"
  check_contains "AGENTS.md" '架构文档规范化流程' "AGENTS references architecture docs normalization workflow"
  check_contains "docs/文档规范.md" '架构文档规范化流程' "documentation guide defines architecture docs normalization workflow"
  check_contains "AGENTS.md" '架构图生成规范' "AGENTS references architecture diagram generation workflow"
  check_contains "docs/文档规范.md" '架构图生成规范' "documentation guide defines architecture diagram generation workflow"
  check_contains "docs/文档规范.md" '每个组件都要写“负责 / 不负责”' "documentation guide requires component responsibility boundaries"
  check_contains "docs/文档规范.md" '每条主流程都要能对应代码路径或 API' "documentation guide requires flow evidence"
  check_contains "docs/文档规范.md" '每个状态都要写触发条件和退出条件' "documentation guide requires state transitions"
  check_contains "docs/文档规范.md" '每个副作用都要写失败后的处理' "documentation guide requires side-effect failure handling"
  check_contains "docs/文档规范.md" '如果写了“支持”，必须说明入口、数据结构、状态变化和测试' "documentation guide constrains support claims"
  check_contains "docs/文档规范.md" '如果不知道，就写 `待确认`' "documentation guide requires explicit unknowns"
  check_contains "docs/文档规范.md" '`当前设计` 质量检查' "documentation guide defines current design quality checks"
  check_contains "scripts/check-open-todos.sh" 'harness/checks/check_open_todos\.sh' "open todos wrapper delegates to harness check"
  check_contains "scripts/check-docs-consistency.py" 'harness/checks/check_docs_consistency\.py' "documentation consistency wrapper delegates to harness check"
  check_contains "harness/checks/check_docs_consistency.py" 'VAGUE_ARCHITECTURE_PHRASES' "documentation consistency script checks vague architecture phrases"
  check_contains "harness/checks/check_docs_consistency.py" 'check_architecture_doc_quality' "documentation consistency script checks architecture doc quality"
  check_contains "AGENTS.md" 'docs/architecture/' "AGENTS references docs/architecture"
  check_contains "AGENTS.md" 'docs/contracts/' "AGENTS references docs/contracts"
  check_contains "AGENTS.md" 'docs/contracts/openapi-v1/' "AGENTS references split OpenAPI source"
  check_contains "AGENTS.md" 'docs/design/' "AGENTS references docs/design"
  check_contains "AGENTS.md" 'docs/plan/impl-plan/' "AGENTS references docs/plan/impl-plan"
  check_contains "AGENTS.md" 'docs/plan/archive/impl-plan/' "AGENTS references docs/plan/archive/impl-plan"
  check_contains "AGENTS.md" 'docs/reviews/' "AGENTS references docs/reviews"
  check_contains "AGENTS.md" 'docs/todos/' "AGENTS references docs/todos"
  check_contains "AGENTS.md" 'scripts/check-task-intake\.sh' "AGENTS references task intake gate"
  check_contains "AGENTS.md" 'scripts/start-implementation\.sh' "AGENTS references implementation startup gate"
  check_contains "AGENTS.md" 'scripts/check-open-todos\.sh' "AGENTS references todo reminder script"
  check_contains "AGENTS.md" 'docs/operations/' "AGENTS references docs/operations"
  check_file "docs/plan/README.md"
  check_file "docs/plan/impl-plan/README.md"
  check_file "docs/plan/archive/impl-plan/README.md"
  check_contains "docs/README.md" 'docs/plan/README\.md' "docs README references plan index"

  echo "[C8a] open todos are surfaced to the operator"
  if [[ -x "scripts/check-open-todos.sh" ]]; then
    bash scripts/check-open-todos.sh --quiet-if-empty
  else
    echo "  $(red FAIL) — scripts/check-open-todos.sh is not executable"
    fail=1
  fi

  echo "[C8b] documentation references stay current"
  python3 scripts/check-docs-consistency.py

  echo "[C8c] OpenAPI source and bundle stay synced"
  python3 scripts/sync_openapi_from_contract.py --check

  echo "[C9] works index covers harness good practices"
  check_file "works/harness-good-practices.md"
  check_contains "works/AGENTS.md" 'harness-good-practices\.md' "works AGENTS references harness good practices"
}
