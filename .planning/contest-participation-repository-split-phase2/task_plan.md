# Task Plan

## Goal

继续收口 `contest` infrastructure，把 `infrastructure/participation_repository.go` 从单文件拆成更清晰的 registration、announcement 与 progress 结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 participation repository 职责 | completed | 已确认单文件同时承载 registration、announcement 与 solved progress 查询 |
| 2. 拆分文件结构 | completed | 已拆为 `participation_repository.go`、`participation_registration_repository.go`、`participation_announcement_repository.go`、`participation_progress_repository.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 定向测试 |

## Acceptance Checks

- `participation_repository.go` 不再承载混杂 ParticipationRepository 逻辑
- registration / announcement / progress 查询拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 `ParticipationRepository` 对外接口
- 仅改善 `contest` infrastructure 文件边界，为后续继续收口 participation 仓储留下更清晰结构
