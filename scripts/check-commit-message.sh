#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "[commit-msg] 用法: bash scripts/check-commit-message.sh <commit-message-file>" >&2
  exit 1
fi

message_file="$1"
if [[ ! -f "$message_file" ]]; then
  echo "[commit-msg] 找不到提交信息文件: $message_file" >&2
  exit 1
fi

subject="$(sed -n '1p' "$message_file" | tr -d '\r')"

if [[ -z "$subject" ]]; then
  echo "[commit-msg] 提交信息不能为空" >&2
  exit 1
fi

# Allow Git-generated merge commits and canonical revert commits.
if [[ "$subject" =~ ^Merge[[:space:]] ]] || [[ "$subject" =~ ^Revert[[:space:]] ]]; then
  exit 0
fi

pattern='^(feat|fix|refactor|docs|test|chore|build|ci|perf|style|revert)(\([^)]+\))?: .+$'
if [[ ! "$subject" =~ $pattern ]]; then
  cat >&2 <<'EOF'
[commit-msg] 提交信息格式不符合约束。
要求：英文类型 + 可选英文/模块 scope + 中文描述
示例：
  fix(frontend): 修正拓扑页导出按钮禁用态
  refactor(topology): 拆分画布工作区组件
  docs: 补齐提交信息约束说明
EOF
  exit 1
fi

description="${subject#*: }"
if ! python3 - "$description" <<'PY'
import re
import sys

description = sys.argv[1]
sys.exit(0 if re.search(r'[\u4e00-\u9fff]', description) else 1)
PY
then
  cat >&2 <<'EOF'
[commit-msg] 提交描述必须包含中文说明。
示例：
  fix(frontend): 修正拓扑页导出按钮禁用态
EOF
  exit 1
fi
