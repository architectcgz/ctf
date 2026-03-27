# Task Plan

## Goal

继续收口 `contest` AWD query 层，把 `application/queries/awd_query.go` 从单文件拆成更清晰的 list 与 summary 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD query 职责 | completed | 已确认单文件同时承载 round/services/attack list 与 round summary |
| 2. 拆分文件结构 | completed | 已拆为 `awd_list_query.go` 与 `awd_summary_query.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_query.go` 不再混载 list 与 summary 主流程
- round/services/attack list 与 round summary 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService` 对外类型与构造函数
- 仅改善 `contest` AWD query 文件边界，为后续继续收口 AWD 查询主流程留下更清晰结构
