# Task Plan

## Goal

继续收口 `contest` 更新命令，把 `application/commands/contest_update_commands.go` 从单文件拆成入口编排与更新校验/字段应用 support 两段，保持 `ContestService.UpdateContest` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 contest update command 职责 | completed | 已确认单文件同时承载资源加载、状态/时间校验与字段应用 |
| 2. 拆分文件结构 | completed | 已拆为 `contest_update_support.go`，`contest_update_commands.go` 保留入口编排与持久化 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `contest_update_commands.go` 不再混载校验与字段应用细节
- 更新前资源加载拆到 support 文件
- 状态/时间校验与字段应用拆到 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ContestService.UpdateContest` 对外接口与状态机行为
- 仅改善 contest update command 文件边界，为后续继续收口 contest commands 留下更清晰结构
