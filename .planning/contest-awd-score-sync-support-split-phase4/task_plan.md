# Task Plan

## Goal

继续收口 `contest` AWD infrastructure support，把 `infrastructure/awd_score_sync_support.go` 从单文件拆成 source 归一化与时间解析两段，保持 helper 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd score sync support 职责 | completed | 已确认单文件同时承载 check/attack source 归一化与 time parse |
| 2. 拆分文件结构 | completed | 已拆为 `awd_score_source_support.go` 与 `awd_score_time_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_score_sync_support.go` 不再混载 source 与 time helper
- check/attack source 归一化 helper 下沉到 `awd_score_source_support.go`
- time parse helper 下沉到 `awd_score_time_support.go`
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 AWD score sync/recalc helper 函数签名与行为
- 仅改善 AWD infrastructure support 文件边界，为后续继续收口 contest infrastructure 留下更清晰结构
