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

body_validation_result="$(
  python3 - "$message_file" <<'PY'
from pathlib import Path
import sys

message_file = Path(sys.argv[1])
lines = message_file.read_text(encoding="utf-8", errors="ignore").splitlines()

body_lines = []
for raw_line in lines[1:]:
    line = raw_line.rstrip("\r")
    if line.startswith("#"):
        continue
    stripped = line.strip()
    if not stripped:
        continue
    body_lines.append(stripped)

visible_chars = sum(
    len("".join(ch for ch in line if not ch.isspace()))
    for line in body_lines
)

if len(body_lines) < 2:
    print("missing_detail_lines")
    sys.exit(1)

if visible_chars < 20:
    print("detail_too_short")
    sys.exit(2)
PY
)" || body_validation_status=$?

body_validation_status="${body_validation_status:-0}"

if [[ "$body_validation_status" -ne 0 ]]; then
  case "$body_validation_result" in
    missing_detail_lines)
      cat >&2 <<'EOF'
[commit-msg] 普通提交不能只有简短标题，必须补充详细正文。
要求：
  - 标题后保留空行
  - 正文至少两行有效内容
  - 建议直接使用多个 -m 组织提交信息
示例：
  git commit -m "feat(frontend): 收口用户页迁移" \
    -m "拆分页面流程 owner，避免 platform 目录继续混放页面逻辑。" \
    -m "同步更新路由接线和测试，方便后续继续迁移 instance。"
EOF
      ;;
    detail_too_short)
      cat >&2 <<'EOF'
[commit-msg] 提交正文信息量不足，请补充更具体的变更说明。
要求：
  - 正文至少两行有效内容
  - 正文总信息量至少达到 20 个非空白字符
示例：
  git commit -m "docs(workflow): 收紧提交说明校验" \
    -m "要求普通提交必须携带详细正文，不能只写单行标题。" \
    -m "同步更新 hook 说明和仓库约定，减少后续提交漂移。"
EOF
      ;;
    *)
      echo "[commit-msg] 提交正文校验失败，请检查提交信息格式" >&2
      ;;
  esac
  exit 1
fi
