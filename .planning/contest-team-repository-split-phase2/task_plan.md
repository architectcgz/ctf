# Task Plan

## Goal

继续收口 `contest` infrastructure，把 `infrastructure/team_repository.go` 从单文件拆成更清晰的成员事务、查询与约束 helper 结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team repository 职责 | completed | 已确认单文件同时承载 team 成员事务、查询与唯一约束判断/helper |
| 2. 拆分文件结构 | completed | 已拆为 `team_repository.go`、`team_membership_repository.go`、`team_query_repository.go`、`team_repository_support.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 与 `team_repository_test.go` 定向测试 |

## Acceptance Checks

- `team_repository.go` 不再承载混杂 TeamRepository 逻辑
- 成员事务 / 查询 / 约束 helper 拆到独立文件
- `contest/...` 与相关 repository 测试通过

## Result

- 不改 `TeamRepository` 对外接口
- 仅改善 `contest` infrastructure 文件边界，为后续继续收口 team 相关仓储留下更清晰结构
