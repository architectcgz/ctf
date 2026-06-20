# CTF Project Hooks

本目录包含项目级别的 hook 脚本，用于在 Agent 执行特定操作前后自动触发检查或拦截。

## 已配置的 Hooks

### PreToolUse Hook: `protect-core-files.sh`

**作用**：拦截对核心规则文件的意外修改，防止 Agent 在"用户着急"或"Leader 要求"等场景下绕过规则。

**触发时机**：Agent 尝试执行 `Edit` 或 `Write` 工具时

**保护的文件**：
- `AGENTS.md` — 项目入口和薄壳路由表
- `harness/policies/reuse-first.yaml` — 复用优先策略
- `harness/policies/commit-message.json` — 提交信息策略
- `harness/policies/project-patterns.yaml` — 项目模式索引
- `harness/policies/script-layer-manifest.json` — 脚本层清单
- `docs/文档规范.md` — 文档规范
- `.agents/skills/*/SKILL.md` — 项目 skill 定义

**允许修改的例外**：
如果提交信息或任务描述中包含以下模式，允许修改：
- `chore(harness): update policy`
- `docs(harness): fix policy typo`
- `feat(harness): extend policy`

**配置位置**：
- Codex: `.codex/hooks.json`
- Claude Code: `.claude/settings.local.json`

## 为什么需要这个 Hook？

根据 [如何写一个好的 skill](https://linux.do/t/topic/1923706) 文章：

> **根因**：模型的注意力方向是"回答用户的请求"，不是"遵守规则"。当用户说"我 leader 让我加"或"demo 5 分钟后要用"，模型更倾向于帮用户，而不是帮规则。

PreToolUse hook 实现**结构性拦截**：

1. **拦截发生在工具调用之前**：Agent 根本没机会写到文件里
2. **Agent 只看到拒绝错误**：无论给多少理由，都不行
3. **必须显式例外**：只有在提交信息中明确说明，才能通过

## 配置说明

### Codex (.codex/hooks.json)

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash harness/hooks/protect-core-files.sh",
            "statusMessage": "Checking protected files"
          }
        ]
      }
    ]
  }
}
```

### Claude Code (.claude/settings.local.json)

```json
{
  "hooks": {
    "preToolUse": [
      {
        "tools": ["edit", "write"],
        "script": "bash harness/hooks/protect-core-files.sh"
      }
    ]
  }
}
```

## 测试 Hook

```bash
# 测试拦截（应该失败）
bash harness/hooks/protect-core-files.sh <<< '{"file_path": "AGENTS.md"}'

# 测试放行（应该成功）
bash harness/hooks/protect-core-files.sh <<< '{"file_path": "code/backend/main.go"}'
```

## 扩展保护列表

如果需要保护更多文件，编辑 `protect-core-files.sh` 中的 `PROTECTED_PATHS` 数组：

```bash
PROTECTED_PATHS=(
  "AGENTS.md"
  "harness/policies/*.yaml"
  "your/new/protected/path.md"
)
```

## 相关文档

- [薄壳设计](../../AGENTS.md#quick-routing--task-entry-points) — Quick Routing 表在 AGENTS.md 顶部
- [Red Flags](../../AGENTS.md#red-flags--stop) — 常见合理化借口和现实
- [Auto-Triggers](../../AGENTS.md#auto-triggers--session-discipline) — 自动触发规则

## 未来可能添加的 Hooks

### PostToolUse Hook (AAR)

任务完成后自动触发 After Action Review：

- 检查：是否有新的坑？
- 更新：Known Gotchas
- 验证：规则文件是否被意外修改？

### PreCommit Hook

提交前的最终检查：

- 运行 `bash scripts/check-commit-message.sh`
- 运行 `bash scripts/check-frontend-test-guard.sh --staged`
- 确保 task slug 已绑定（非琐碎任务）
