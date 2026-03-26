# Task Plan

## Goal

继续收口 `contest` HTTP 层，把 `api/http/participation_handler.go` 从单文件拆成更清晰的 command 与 query handler 结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 participation handler 职责 | completed | 已确认单文件同时承载 registration/announcement 命令与查询入口 |
| 2. 拆分文件结构 | completed | 已拆为 `participation_handler.go`、`participation_command_handler.go`、`participation_query_handler.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 与相关 `internal/app` 定向测试 |

## Acceptance Checks

- `participation_handler.go` 不再承载混杂 ParticipationHandler HTTP 逻辑
- command / query HTTP 入口拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ParticipationHandler` 对外类型与构造函数
- 仅改善 `contest` HTTP 文件边界，为后续继续收口 participation 相关入口留下更清晰结构
