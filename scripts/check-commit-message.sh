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

root_dir="$(git rev-parse --show-toplevel)"
agents_home="${AGENTS_HOME:-$HOME/.agents}"
checker="$agents_home/harness/commit-message/check_commit_message.py"
policy_file="$root_dir/harness/policies/commit-message.json"

if [[ ! -f "$checker" ]]; then
  echo "[commit-msg] 找不到共享提交信息检查器: $checker" >&2
  echo "[commit-msg] 请先同步 ~/.agents/harness，或检查 AGENTS_HOME 配置" >&2
  exit 1
fi

if [[ ! -f "$policy_file" ]]; then
  echo "[commit-msg] 找不到项目提交信息策略: $policy_file" >&2
  exit 1
fi

cmd=(python3 "$checker" --message-file "$message_file" --policy-file "$policy_file")
if [[ -x "$root_dir/scripts/check-startup-gate.sh" ]]; then
  active_task_slug="$(bash "$root_dir/scripts/check-startup-gate.sh" --print-active-slug 2>/dev/null || true)"
  if [[ -n "$active_task_slug" ]]; then
    cmd+=(--active-task-slug "$active_task_slug")
  fi
fi

exec "${cmd[@]}"
