# Task Plan

## Goal

继续收口 `contest` infrastructure，把 `infrastructure/contest_repository.go` 从单文件拆成基础 CRUD/list 与状态推进仓储两段，保持 `Repository` 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 contest repository 职责 | completed | 已确认单文件同时承载基础 CRUD/list 与状态推进查询/更新 |
| 2. 拆分文件结构 | completed | 已保留 `contest_repository.go` 承载基础 CRUD/list，并新增 `contest_status_repository.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `contest_repository.go` 不再混载基础 CRUD/list 与状态推进查询/更新
- 状态推进查询与更新拆到独立 repository 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `Repository` 对外接口与 contest 状态推进行为
- 仅改善 `contest` infrastructure 文件边界，为后续继续收口 contest repository 链路留下更清晰结构
