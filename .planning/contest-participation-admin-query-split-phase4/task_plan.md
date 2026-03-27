# Task Plan

## Goal

继续收口 `contest` participation queries，把 `application/queries/participation_admin_query.go` 从单文件拆成 registration admin query、announcement query 与 contest existence support 三段，保持 `ParticipationService` admin 查询行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 participation admin query 职责 | completed | 已确认单文件同时承载 registration admin query 与 announcement query |
| 2. 拆分文件结构 | completed | 已拆为 `participation_registration_admin_query.go`、`participation_announcement_query.go` 与 `participation_query_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `participation_admin_query.go` 的两类 admin query 被拆到独立文件
- registration list query 拆到独立文件
- announcement list query 拆到独立文件
- 共享 contest existence 校验收口到 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ParticipationService` admin 查询接口与返回结构
- 仅改善 participation query 文件边界，为后续继续收口 participation application queries 留下更清晰结构
