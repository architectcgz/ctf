# Task Plan

## Goal

继续收口 `contest` query 层，把 `application/queries/challenge_query.go` 从单文件拆成更清晰的 admin 与 visible 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 challenge query 职责 | completed | 已确认单文件同时承载 admin 与 visible challenge 查询 |
| 2. 拆分文件结构 | completed | 已拆为 `challenge_admin_query.go` 与 `challenge_visible_query.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `challenge_query.go` 不再混载 admin 与 visible challenge 查询
- admin 与 visible challenge 查询拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ChallengeService` 对外类型与构造函数
- 仅改善 `contest` challenge query 文件边界，为后续继续收口 query 主流程留下更清晰结构
