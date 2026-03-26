# Task Plan

## Goal

继续收口 `contest` AWD infrastructure，把 `awd_repository.go` 从单文件拆成更清晰的仓储职责文件，保持对外接口不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWDRepository 职责 | completed | 已确认单文件同时承载基础仓储、round、contest/team/challenge 关系、service/attack 记录能力 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_repository.go`、`awd_round_repository.go`、`awd_relation_repository.go`、`awd_service_repository.go` |
| 3. focused 验证 | completed | 已运行 `./internal/module/contest/...` 定向测试 |

## Acceptance Checks

- `awd_repository.go` 仅保留基础仓储定义与共享 helper
- round、关系查询、service/attack 记录能力拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 `AWDRepository` 对外接口
- 仅改善 infrastructure 层文件边界，为后续继续拆 AWD 基础设施留出清晰落点
