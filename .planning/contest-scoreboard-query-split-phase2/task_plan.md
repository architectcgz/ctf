# Task Plan

## Goal

继续收口 `contest` query 层，把 `application/queries/scoreboard_query.go` 从单文件拆成更清晰的 scoreboard list 与 team rank 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 scoreboard query 职责 | completed | 已确认单文件同时承载榜单分页查询与 team rank 查询 |
| 2. 拆分文件结构 | completed | 已拆为 `scoreboard_list_query.go` 与 `scoreboard_rank_query.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `scoreboard_query.go` 不再混载 scoreboard list 与 team rank 查询
- scoreboard list 与 team rank 查询拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ScoreboardService` 对外类型与构造函数
- 仅改善 `contest` scoreboard query 文件边界，为后续继续收口 query 主流程留下更清晰结构
