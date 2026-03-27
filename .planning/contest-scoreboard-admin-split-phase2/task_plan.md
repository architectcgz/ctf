# Task Plan

## Goal

继续收口 `contest` command 层，把 `application/commands/scoreboard_admin_service.go` 从单文件拆成更清晰的 score 与 freeze 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 scoreboard admin 职责 | completed | 已确认单文件同时承载 score rebuild 与 freeze/unfreeze 流程 |
| 2. 拆分文件结构 | completed | 已拆为 `scoreboard_admin_service.go`、`scoreboard_admin_score_commands.go`、`scoreboard_admin_freeze_commands.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `scoreboard_admin_service.go` 不再混载 score 与 freeze 具体流程
- score rebuild 与 freeze/unfreeze 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ScoreboardAdminService` 对外类型与构造函数
- 仅改善 `contest` scoreboard admin command 文件边界，为后续继续收口 command 主流程留下更清晰结构
