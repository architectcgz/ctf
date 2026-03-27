# Task Plan

## Goal

继续收口 `contest` AWD query，把 `application/queries/awd_list_query.go` 从单文件拆成 round list、service list、attack log list 三段，保持 `AWDService` 对外查询行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd list query 职责 | completed | 已确认单文件同时承载 round 列表、service 列表、attack log 列表三类查询 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_round_list_query.go`、`awd_service_list_query.go`、`awd_attack_log_list_query.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_list_query.go` 不再混载 round/service/attack log 三类查询
- 三类 AWD list query 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService` 对外查询接口与返回结构
- 仅改善 `contest` AWD query 文件边界，为后续继续收口 AWD read-side 链路留下更清晰结构
