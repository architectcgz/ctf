# Task Plan

## Goal

继续收口 `contest` HTTP handlers，把 `api/http/scoreboard_handler.go` 从单文件拆成 scoreboard query 与 freeze/unfreeze admin 两段，保持 `Handler` 对外路由与响应行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 scoreboard handler 职责 | completed | 已确认单文件同时承载 scoreboard query 与 freeze/unfreeze admin 两类 HTTP 入口 |
| 2. 拆分文件结构 | completed | 已拆为 `scoreboard_query_handler.go` 与 `scoreboard_admin_handler.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `scoreboard_handler.go` 不再混载 scoreboard query 与 freeze/unfreeze admin 两类 HTTP 入口
- scoreboard query 与 admin handler 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `Handler` 对外路由、参数校验与响应行为
- 仅改善 `contest` scoreboard HTTP 文件边界，为后续继续收口 scoreboard handler 链路留下更清晰结构
