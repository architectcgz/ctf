#!/usr/bin/env bash
#
# PreToolUse hook: 保护核心规则文件，防止意外修改
#
# 用法：
#   在 .codex/hooks.json 或 .claude/settings.json 中配置：
#   {
#     "PreToolUse": {
#       "Edit|Write": {
#         "command": "bash harness/hooks/protect-core-files.sh",
#         "description": "Protect core policy files from accidental modification"
#       }
#     }
#   }
#
# 环境变量（由 hook 框架注入）：
#   TOOL_NAME - 工具名称（Edit / Write）
#   TOOL_ARGS_JSON - 工具参数的 JSON 表示
#
set -euo pipefail

# 受保护的文件列表（相对于项目根）
PROTECTED_PATHS=(
  "AGENTS.md"
  "harness/policies/reuse-first.yaml"
  "harness/policies/commit-message.json"
  "harness/policies/project-patterns.yaml"
  "harness/policies/script-layer-manifest.json"
  "docs/文档规范.md"
  ".agents/skills/*/SKILL.md"
)

# 允许修改的例外情况（需要在提交信息或任务描述中明确说明）
ALLOW_PATTERNS=(
  "chore(harness): update policy"
  "docs(harness): fix policy typo"
  "feat(harness): extend policy"
)

# 从 TOOL_ARGS_JSON 中提取 file_path
extract_file_path() {
  local json="$1"
  # 简单的 JSON 解析，提取 file_path 字段
  echo "$json" | grep -oP '"file_path"\s*:\s*"\K[^"]+' || echo ""
}

# 检查路径是否匹配保护列表
is_protected() {
  local path="$1"
  local protected_pattern

  for protected_pattern in "${PROTECTED_PATHS[@]}"; do
    # 支持通配符匹配
    if [[ "$path" == $protected_pattern ]]; then
      return 0
    fi

    # 处理绝对路径（转为相对路径）
    local rel_path="${path#$PWD/}"
    if [[ "$rel_path" == $protected_pattern ]]; then
      return 0
    fi
  done

  return 1
}

# 检查是否在允许的例外模式中
is_allowed_exception() {
  local reason="$1"
  local pattern

  for pattern in "${ALLOW_PATTERNS[@]}"; do
    if [[ "$reason" =~ $pattern ]]; then
      return 0
    fi
  done

  return 1
}

main() {
  # 获取工具参数
  local tool_name="${TOOL_NAME:-}"
  local tool_args="${TOOL_ARGS_JSON:-}"

  # 如果没有注入环境变量，尝试从命令行参数读取
  if [[ -z "$tool_args" ]]; then
    # Codex 可能通过不同方式传参，这里提供兼容处理
    if [[ $# -gt 0 ]]; then
      tool_args="$1"
    else
      # 无法获取参数，放行（避免误拦）
      exit 0
    fi
  fi

  # 提取文件路径
  local file_path
  file_path=$(extract_file_path "$tool_args")

  if [[ -z "$file_path" ]]; then
    # 无法提取路径，放行
    exit 0
  fi

  # 检查是否受保护
  if is_protected "$file_path"; then
    # 检查是否有例外说明
    local task_context="${TASK_CONTEXT:-}"
    local commit_message="${COMMIT_MESSAGE:-}"

    if is_allowed_exception "$task_context" || is_allowed_exception "$commit_message"; then
      # 允许例外修改
      echo "[PreToolUse] Allowing protected file modification: $file_path (exception matched)" >&2
      exit 0
    fi

    # 拦截修改
    cat >&2 <<EOF

[PreToolUse Hook] BLOCKED: Protected file modification attempt
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

File: $file_path

This file is protected by PreToolUse hook. Direct modification is not allowed.

Why this matters:
- Core policy files define project rules and should not be casually modified
- Accidental edits to these files can break the entire harness
- Changes to these files need explicit review and approval

If you really need to modify this file:
1. Confirm with the user that this change is intentional
2. Explain why the policy needs to change
3. Document the change in feedback/ or commit message
4. Use an allowed commit pattern:
   - chore(harness): update policy
   - docs(harness): fix policy typo
   - feat(harness): extend policy

Protected files list:
$(printf '  - %s\n' "${PROTECTED_PATHS[@]}")

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EOF

    # 返回非零状态码，拦截操作
    exit 1
  fi

  # 不在保护列表中，放行
  exit 0
}

main "$@"
