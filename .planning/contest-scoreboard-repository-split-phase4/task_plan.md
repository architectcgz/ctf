# Task Plan

## Goal

继续收口 `contest` infrastructure，把 `infrastructure/contest_scoreboard_repository.go` 从单文件拆成公共入口、mode-specific 查询分支与 aggregate time parsing support 三段，保持 scoreboard team stats 查询行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 contest scoreboard repository 职责 | completed | 已确认单文件同时承载 AWD/非 AWD 查询分支与 aggregate time parsing support |
| 2. 拆分文件结构 | completed | 已拆为 `contest_scoreboard_mode_repository.go` 与 `contest_scoreboard_time_support.go`，`contest_scoreboard_repository.go` 保留入口与结果映射 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `contest_scoreboard_repository.go` 不再混载 mode-specific 查询分支与 time parsing support
- AWD/非 AWD scoreboard stats 查询拆到独立 mode repository 文件
- aggregate time parsing support 拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `Repository.FindScoreboardTeamStats` 对外接口与返回结构
- 仅改善 contest scoreboard repository 文件边界，为后续继续收口 scoreboard repository / query 链路留下更清晰结构
