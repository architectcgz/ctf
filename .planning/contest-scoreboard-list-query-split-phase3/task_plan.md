# Task Plan

## Goal

继续收口 `contest` query，把 `application/queries/scoreboard_list_query.go` 从单文件拆成 scoreboard 查询主流程与分页/key/item 组装 support 两段，保持返回结构与榜单行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 scoreboard list query 职责 | completed | 已确认单文件同时承载 scoreboard 查询入口与分页/key/item 组装 support |
| 2. 拆分文件结构 | completed | 已保留 `scoreboard_list_query.go` 为主流程，并新增 `scoreboard_list_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `scoreboard_list_query.go` 不再混载查询入口与分页/key/item 组装 support
- scoreboard support 逻辑拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ScoreboardService` 对外查询行为、冻结榜切换与返回结构
- 仅改善 `contest` scoreboard query 文件边界，为后续继续收口 read-side 链路留下更清晰结构
