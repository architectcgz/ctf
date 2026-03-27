# Task Plan

## Goal

继续收口 `contest` AWD infrastructure，把 `infrastructure/awd_service_repository.go` 从单文件拆成更清晰的 service instance、team service、attack log 三段，保持 `AWDRepository` 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd service repository 职责 | completed | 已确认单文件同时承载 service instance 查询、team service 持久化、attack log/impact 持久化 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_service_instance_repository.go`、`awd_team_service_repository.go`、`awd_attack_log_repository.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_service_repository.go` 不再混载三类 AWD infrastructure 职责
- service instance 查询、team service 持久化、attack log/impact 持久化拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRepository` 对外接口
- 仅改善 `contest` AWD infrastructure 文件边界，为后续继续收口 AWD repository 链路留下更清晰结构
