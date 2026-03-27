# Task Plan

## Goal

继续收口 `contest` query 层，把 `application/queries/contest_service.go` 从单文件拆成更清晰的 get 与 list 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 contest query 职责 | completed | 已确认单文件主要承载 get 与 list 两类查询 |
| 2. 拆分文件结构 | completed | 已拆为 `contest_service.go`、`contest_get_query.go`、`contest_list_query.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `contest_service.go` 不再混载 get/list 具体查询流程
- get 与 list 查询拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ContestService` 对外类型与构造函数
- 仅改善 `contest` query 文件边界，为后续继续收口 query 主流程留下更清晰结构
