# Task Plan

## Goal

继续收口 `contest` AWD HTTP 主流程，把 `api/http/awd_handler.go` 从单文件拆成更清晰的 route handler 分段结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD handler 职责 | completed | 已确认单文件同时承载 round、service check、attack 与 query summary 四类 HTTP 入口 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_handler.go`、`awd_round_handler.go`、`awd_service_handler.go`、`awd_attack_handler.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 与相关 `internal/app` 路由测试 |

## Acceptance Checks

- `awd_handler.go` 不再承载混杂 AWD HTTP 流程
- round / service check / attack HTTP 入口拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDHandler` 对外类型与构造函数
- 仅改善 AWD HTTP 主流程文件边界，为后续继续收口 contest 组合与路由层留下更清晰结构
