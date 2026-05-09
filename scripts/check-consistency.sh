#!/usr/bin/env bash
set -euo pipefail

fail=0

red() { printf '\033[31m%s\033[0m' "$1"; }
green() { printf '\033[32m%s\033[0m' "$1"; }

check_file() {
  if [[ -f "$1" ]]; then
    echo "  $(green PASS) — $1"
  else
    echo "  $(red FAIL) — missing $1"
    fail=1
  fi
}

check_dir() {
  if [[ -d "$1" ]]; then
    echo "  $(green PASS) — $1"
  else
    echo "  $(red FAIL) — missing $1"
    fail=1
  fi
}

check_contains() {
  local file="$1" pattern="$2" label="$3"
  if [[ ! -f "$file" ]]; then
    echo "  $(red FAIL) — $label: missing $file"
    fail=1
  elif grep -qE "$pattern" "$file"; then
    echo "  $(green PASS) — $label"
  else
    echo "  $(red FAIL) — $label"
    fail=1
  fi
}

echo "[C1] strict harness directories exist"
for dir in concepts thinking practice feedback works prompts references; do
  check_dir "$dir"
  check_file "$dir/AGENTS.md"
done

echo "[C2] root navigation references strict harness"
check_contains "AGENTS.md" 'concepts/' "AGENTS references concepts"
check_contains "AGENTS.md" 'thinking/' "AGENTS references thinking"
check_contains "AGENTS.md" 'practice/' "AGENTS references practice"
check_contains "AGENTS.md" 'feedback/' "AGENTS references feedback"
check_contains "AGENTS.md" 'works/' "AGENTS references works"
check_contains "AGENTS.md" 'prompts/' "AGENTS references prompts"
check_contains "AGENTS.md" 'references/' "AGENTS references references"

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
  check_contains ".githooks/pre-commit" 'scripts/check-consistency\.sh' "pre-commit runs scripts/check-consistency.sh"
else
  echo "  $(red FAIL) — missing .githooks/pre-commit"
  fail=1
fi

echo "[C6] migrated accumulation indexes exist"
check_file "feedback/improvements-index.md"
check_file "practice/superpowers-plan-index.md"
check_file "practice/planning-archive-index.md"
check_file "prompts/ctf-ui-theme-system-skill.md"
check_file "prompts/harness-router.md"
check_file "references/ctf-instance-lifecycle-research.md"
check_file "works/harness-migration-map.md"

echo "[C7] migrated indexes are discoverable from directory AGENTS"
check_contains "feedback/AGENTS.md" 'improvements-index\.md' "feedback AGENTS references migrated improvements"
check_contains "practice/AGENTS.md" 'superpowers-plan-index\.md' "practice AGENTS references superpowers index"
check_contains "practice/AGENTS.md" 'planning-archive-index\.md' "practice AGENTS references planning archive"
check_contains "prompts/AGENTS.md" 'ctf-ui-theme-system-skill\.md' "prompts AGENTS references UI theme skill"
check_contains "prompts/AGENTS.md" 'harness-router\.md' "prompts AGENTS references harness router"
check_contains "prompts/AGENTS.md" 'architecture-diagram-generation\.md' "prompts AGENTS references architecture diagram prompt"
check_contains "references/AGENTS.md" 'ctf-instance-lifecycle-research\.md' "references AGENTS references lifecycle research"
check_contains "works/AGENTS.md" 'harness-migration-map\.md' "works AGENTS references migration map"

echo "[C8] AGENTS captures file placement rules"
check_file "docs/文档规范.md"
check_contains "AGENTS.md" 'docs/文档规范\.md' "AGENTS references documentation guide"
check_contains "docs/README.md" 'docs/文档规范\.md' "docs README references documentation guide"
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
check_contains "scripts/check-docs-consistency.py" 'VAGUE_ARCHITECTURE_PHRASES' "documentation consistency script checks vague architecture phrases"
check_contains "scripts/check-docs-consistency.py" 'check_architecture_doc_quality' "documentation consistency script checks architecture doc quality"
check_contains "AGENTS.md" 'docs/architecture/' "AGENTS references docs/architecture"
check_contains "AGENTS.md" 'docs/contracts/' "AGENTS references docs/contracts"
check_contains "AGENTS.md" 'docs/design/' "AGENTS references docs/design"
check_contains "AGENTS.md" 'docs/plan/impl-plan/' "AGENTS references docs/plan/impl-plan"
check_contains "AGENTS.md" 'docs/reviews/' "AGENTS references docs/reviews"
check_contains "AGENTS.md" 'docs/todos/' "AGENTS references docs/todos"
check_contains "AGENTS.md" 'docs/operations/' "AGENTS references docs/operations"

echo "[C8b] documentation references stay current"
python3 scripts/check-docs-consistency.py

echo "[C9] works index covers harness good practices"
check_file "works/harness-good-practices.md"
check_contains "works/AGENTS.md" 'harness-good-practices\.md' "works AGENTS references harness good practices"

echo "[C10] local architecture guardrails are wired"
check_file "scripts/check-architecture.sh"
check_file "scripts/doctor-local-harness.sh"
check_contains ".githooks/pre-commit" 'scripts/check-architecture\.sh --quick' "pre-commit runs quick architecture checks"
check_contains ".githooks/README.md" 'scripts/check-architecture\.sh --quick' "hook docs mention architecture checks"
check_contains "docs/architecture/README.md" 'scripts/check-architecture\.sh --full' "architecture README maps full architecture checks"
check_contains "docs/architecture/README.md" 'code/backend/internal/module/architecture_test\.go' "architecture README references backend guardrail test"
check_contains "docs/architecture/README.md" 'code/frontend/src/__tests__/architectureBoundaries\.test\.ts' "architecture README references frontend guardrail test"

if [[ "$fail" -eq 0 ]]; then
  echo "$(green '✓ all harness consistency checks passed')"
else
  echo "$(red '✗ harness consistency checks failed')"
fi

exit "$fail"
