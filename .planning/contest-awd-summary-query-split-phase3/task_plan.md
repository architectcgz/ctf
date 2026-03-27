# Task Plan

## Goal

继续收口 `contest` AWD query，把 `application/queries/awd_summary_query.go` 从单文件拆成查询主流程与 round summary 汇总 support 两段，保持返回结构与统计口径不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd summary query 职责 | completed | 已确认单文件同时承载 round summary 查询入口与 metrics/items 聚合逻辑 |
| 2. 拆分文件结构 | completed | 已保留 `awd_summary_query.go` 为主流程，并新增 `awd_summary_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_summary_query.go` 不再混载查询入口与 metrics/items 聚合 support
- round summary 聚合逻辑拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService.GetRoundSummary` 对外行为、返回结构与排序口径
- 仅改善 `contest` AWD query 文件边界，为后续继续收口 AWD read-side 链路留下更清晰结构
