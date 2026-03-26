# Task Plan

## Goal

继续收口 `contest` 应用层，把 `application/queries/team_service.go` 从单文件拆成更清晰的 team info 与 team list 查询结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team query 职责 | completed | 已确认单文件同时承载 team info、team list 与 my team 三类查询 |
| 2. 拆分文件结构 | completed | 已拆为 `team_service.go`、`team_info_query.go`、`team_list_query.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 与 `team_service_test.go` 定向测试 |

## Acceptance Checks

- `team_service.go` 不再承载混杂 TeamService 查询逻辑
- team info / team list / my team 查询拆到独立文件
- `contest/...` 与相关 query 测试通过

## Result

- 不改 `TeamService` 对外接口
- 仅改善 `contest` queries 文件边界，为后续继续收口 team 相关查询留下更清晰结构
