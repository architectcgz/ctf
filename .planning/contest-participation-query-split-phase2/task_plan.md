# Task Plan

## Goal

继续收口 `contest` query 层，把 `application/queries/participation_query.go` 从单文件拆成更清晰的 admin query 与 progress query 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 participation query 职责 | completed | 已确认单文件同时承载 registrations/announcements 与 my progress 查询 |
| 2. 拆分文件结构 | completed | 已拆为 `participation_admin_query.go` 与 `participation_progress_query.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `participation_query.go` 不再混载 admin query 与 progress query
- registrations/announcements 与 my progress 查询拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ParticipationService` 对外类型与构造函数
- 仅改善 `contest` participation query 文件边界，为后续继续收口 query 主流程留下更清晰结构
