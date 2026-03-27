# Task Plan

## Goal

继续收口 `contest` AWD HTTP handlers，把 `api/http/awd_round_handler.go` 从单文件拆成 round create/list、round checks、round summary 三段，保持 `AWDHandler` 对外路由与响应行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd round handler 职责 | completed | 已确认单文件同时承载 round create/list、round checks、round summary 三类 HTTP 入口 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_round_manage_handler.go`、`awd_round_check_handler.go`、`awd_round_summary_handler.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_round_handler.go` 不再混载 round create/list、round checks、round summary 三类 HTTP 入口
- 三类 AWD round handler 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDHandler` 对外路由、参数读取与响应行为
- 仅改善 `contest` AWD HTTP 文件边界，为后续继续收口 AWD handler 链路留下更清晰结构
