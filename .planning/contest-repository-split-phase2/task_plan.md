# Task Plan

## Goal

继续收口 `contest` infrastructure，把 `infrastructure/repository.go` 从单文件拆成更清晰的 contest CRUD、team lookup 与 scoreboard 聚合结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 repository 职责 | completed | 已确认单文件同时承载 contest CRUD、状态调度查询、team lookup 与 scoreboard stats 聚合 |
| 2. 拆分文件结构 | completed | 已拆为 `repository.go`、`contest_repository.go`、`contest_team_lookup_repository.go`、`contest_scoreboard_repository.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 与相关 `internal/app` 定向测试 |

## Acceptance Checks

- `repository.go` 不再承载混杂 contest repository 逻辑
- contest CRUD / team lookup / scoreboard 聚合拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `Repository` 对外接口
- 仅改善 `contest` infrastructure 文件边界，为后续继续收口 repository 层留下更清晰结构
