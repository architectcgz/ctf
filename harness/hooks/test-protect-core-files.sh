#!/usr/bin/env bash
#
# 测试 protect-core-files.sh hook 的拦截和放行逻辑
#
set -euo pipefail

HOOK_SCRIPT="harness/hooks/protect-core-files.sh"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_test_header() {
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  $1"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

pass() {
  PASSED_TESTS=$((PASSED_TESTS + 1))
  TOTAL_TESTS=$((TOTAL_TESTS + 1))
  echo -e "  ${GREEN}✓ PASS${NC}: $1"
}

fail() {
  FAILED_TESTS=$((FAILED_TESTS + 1))
  TOTAL_TESTS=$((TOTAL_TESTS + 1))
  echo -e "  ${RED}✗ FAIL${NC}: $1"
  echo -e "    ${RED}$2${NC}"
}

# 测试：应该被拦截的文件
test_should_block() {
  local file_path="$1"
  local test_name="$2"

  local json="{\"file_path\": \"$file_path\"}"

  if TOOL_ARGS_JSON="$json" bash "$HOOK_SCRIPT" 2>/dev/null; then
    fail "$test_name" "Expected block, but allowed: $file_path"
  else
    pass "$test_name"
  fi
}

# 测试：应该被放行的文件
test_should_allow() {
  local file_path="$1"
  local test_name="$2"

  local json="{\"file_path\": \"$file_path\"}"

  if TOOL_ARGS_JSON="$json" bash "$HOOK_SCRIPT" 2>/dev/null; then
    pass "$test_name"
  else
    fail "$test_name" "Expected allow, but blocked: $file_path"
  fi
}

# 主测试套件
main() {
  cd "$(dirname "$0")/../.." || exit 1

  if [[ ! -f "$HOOK_SCRIPT" ]]; then
    echo -e "${RED}Error: Hook script not found: $HOOK_SCRIPT${NC}"
    exit 1
  fi

  echo ""
  echo "Testing protect-core-files.sh hook"

  # 测试组 1: 应该被拦截的受保护文件
  print_test_header "Test Group 1: Protected Files (Should Block)"

  test_should_block "AGENTS.md" "Block AGENTS.md"
  test_should_block "harness/policies/reuse-first.yaml" "Block reuse-first.yaml"
  test_should_block "harness/policies/commit-message.json" "Block commit-message.json"
  test_should_block "docs/文档规范.md" "Block 文档规范.md"
  test_should_block ".agents/skills/ctf-backend-patterns/SKILL.md" "Block project skill SKILL.md"

  # 测试组 2: 应该被放行的非受保护文件
  print_test_header "Test Group 2: Non-Protected Files (Should Allow)"

  test_should_allow "code/backend/main.go" "Allow backend code"
  test_should_allow "code/frontend/src/App.vue" "Allow frontend code"
  test_should_allow "docs/architecture/backend.md" "Allow architecture docs"
  test_should_allow "feedback/踩坑记录.md" "Allow feedback files"
  test_should_allow "harness/checks/some-check.sh" "Allow harness checks"

  # 测试组 3: 边界情况
  print_test_header "Test Group 3: Edge Cases"

  test_should_allow "harness/policies/README.md" "Allow non-policy files in policies dir"
  test_should_allow "docs/README.md" "Allow non-规范 docs"
  test_should_allow ".agents/skills/README.md" "Allow skills README (not SKILL.md)"

  # 总结
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "  Test Summary"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  echo "  Total:  $TOTAL_TESTS"
  echo -e "  ${GREEN}Passed: $PASSED_TESTS${NC}"
  if [[ $FAILED_TESTS -gt 0 ]]; then
    echo -e "  ${RED}Failed: $FAILED_TESTS${NC}"
    echo ""
    exit 1
  else
    echo -e "  ${YELLOW}Failed: 0${NC}"
    echo ""
    echo -e "${GREEN}All tests passed!${NC}"
    echo ""
  fi
}

main "$@"
