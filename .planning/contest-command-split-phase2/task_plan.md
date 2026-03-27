# Task Plan

## Goal

继续收口 `contest` command 层，把 `application/commands/contest_service.go` 从单文件拆成更清晰的 create 与 update 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 contest command 职责 | completed | 已确认单文件主要承载 create 与 update 两类流程 |
| 2. 拆分文件结构 | completed | 已拆为 `contest_service.go`、`contest_create_commands.go`、`contest_update_commands.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `contest_service.go` 不再混载 create/update 具体流程
- create 与 update 命令拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ContestService` 对外类型与构造函数
- 仅改善 `contest` command 文件边界，为后续继续收口 application 主流程留下更清晰结构
