# Task Plan

## Goal

继续收口 `contest` HTTP 层，把 `api/http/handler.go` 从单文件拆成更清晰的 contest CRUD 与 scoreboard handler 结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 handler 职责 | completed | 已确认单文件同时承载 contest CRUD/query 与 scoreboard 入口 |
| 2. 拆分文件结构 | completed | 已拆为 `handler.go`、`contest_handler.go`、`scoreboard_handler.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `handler.go` 不再承载混杂 contest HTTP 逻辑
- contest CRUD/query 与 scoreboard 入口拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `Handler` 对外类型与构造函数
- 仅改善 `contest` HTTP 文件边界，为后续继续收口总入口留下更清晰结构
