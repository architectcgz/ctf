# Task Plan

## Goal

继续收口 `contest` AWD service upsert command，把 `application/commands/awd_service_upsert_commands.go` 从单文件拆成入口编排、事务写入重算与后置同步响应三段，保持 `AWDService.UpsertServiceCheck` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd service upsert command 职责 | completed | 已确认单文件同时承载参数编排、事务写入重算和后置缓存/状态同步 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_service_upsert_transaction.go` 与 `awd_service_upsert_response_support.go`，`awd_service_upsert_commands.go` 保留入口编排 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_service_upsert_commands.go` 不再混载事务写入重算与后置同步
- 事务写入 service check 与 team score 重算拆到独立 transaction 文件
- 缓存重建、服务状态同步与响应映射拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService.UpsertServiceCheck` 对外接口、返回结构与业务行为
- 仅改善 AWD service upsert command 文件边界，为后续继续收口 AWD commands 留下更清晰结构
