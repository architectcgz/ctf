# Task Plan

## Goal

继续收口 `contest` status repository，把 `infrastructure/contest_status_repository.go` 从单文件拆成状态筛选查询与状态更新写入两段，保持 `Repository` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 contest status repository 职责 | completed | 已确认单文件同时承载状态筛选查询与状态更新写入 |
| 2. 拆分文件结构 | completed | 已拆为 `contest_status_repository.go`（list）与 `contest_status_update_repository.go`（update） |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `contest_status_repository.go` 不再混载状态更新写入逻辑
- `UpdateStatus` 与 `contestExists` 拆到独立 update 文件
- `ListByStatusesAndTimeRange` 保持查询逻辑不变
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `Repository` 对外接口与状态推进行为
- 仅改善 contest status repository 文件边界，为后续继续收口 infrastructure 留下更清晰结构
