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
  check_contains "AGENTS.md" 'harness/workflow-plugins/' "AGENTS references workflow plugin adapters"
  check_contains "AGENTS.md" 'harness/prompts/' "AGENTS references harness prompts"
  check_contains "AGENTS.md" 'references/' "AGENTS references references"

  echo "[C2a] project agent entrypoints stay aligned"
  check_file "scripts/check-agent-entrypoints.sh"
  if [[ -x "scripts/check-agent-entrypoints.sh" ]]; then
    bash scripts/check-agent-entrypoints.sh
  else
    echo "  $(red FAIL) — scripts/check-agent-entrypoints.sh is not executable"
    fail=1
  fi

  if [[ -e "harness/bridges" ]]; then
    echo "  $(red FAIL) — harness/bridges must not exist; ordinary project commands should live in scripts/ or harness/checks/"
    fail=1
  else
    echo "  $(green PASS) — harness/bridges is absent"
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
  check_file "scripts/check-workflow-governance.sh"
  check_file "scripts/check-workflow-governance-core.sh"
  check_file "scripts/check-review-governance.sh"
  check_file "scripts/check-review-governance-core.sh"
  check_file "scripts/run-workflow-stage.sh"
  check_file "harness/policies/commit-message.json"
  check_file "scripts/check-frontend-test-guard.sh"
  check_file "scripts/check-startup-gate.sh"
  check_contains ".githooks/commit-msg" 'scripts/check-commit-message\.sh' "commit-msg runs scripts/check-commit-message.sh"
  check_contains "scripts/check-consistency.sh" 'scripts/check-workflow-governance\.sh' "check-consistency delegates to workflow governance"
  check_contains "scripts/check-workflow-governance.sh" 'scripts/run-workflow-stage\.sh workflow-governance' "workflow governance wrapper uses stage runner"
  check_contains "scripts/check-review-governance.sh" 'scripts/check-workflow-governance\.sh' "review governance wrapper stays as compatibility alias"
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
  check_file "scripts/check-agent-entrypoints.sh"
  check_file "scripts/check-docs-consistency.py"
  check_file "scripts/check-script-guard.sh"
  check_file "scripts/check-script-layer.sh"
  check_file "harness/checks/check_open_todos.sh"
  check_file "harness/checks/check_docs_consistency.py"
  check_file "harness/checks/check_script_layer_conventions.py"
  check_file "harness/policies/script-guard.json"
  check_file "harness/policies/script-layer-manifest.json"
  check_file "tools/AGENTS.md"
  check_file "tools/sync_openapi_from_contract.py"
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
  check_contains "scripts/check-script-guard.sh" 'check_script_guard\.py' "script guard wrapper delegates to shared harness check"
  check_contains "scripts/check-script-layer.sh" 'harness/checks/check_script_layer_conventions\.py' "script layer wrapper delegates to harness check"
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
  check_contains "AGENTS.md" 'tools/' "AGENTS references tools layer"
  check_contains "AGENTS.md" 'script-layer-manifest\.json' "AGENTS references script layer manifest"
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
  python3 tools/sync_openapi_from_contract.py --check

  echo "[C8d] script layer naming and ownership stay aligned"
  bash scripts/check-script-layer.sh

  echo "[C8e] script growth stays under guard"
  bash scripts/check-script-guard.sh

  echo "[C9] works index covers harness good practices"
  check_file "works/harness-good-practices.md"
  check_contains "works/AGENTS.md" 'harness-good-practices\.md' "works AGENTS references harness good practices"
}
