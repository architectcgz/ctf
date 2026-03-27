# Task Plan

## Goal

继续收口 `contest` AWD command support，把 `application/commands/awd_round_support.go` 从单文件拆成更清晰的当前轮次解析与 live window 判定两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd round support 职责 | completed | 已确认单文件同时承载当前轮次解析/物化与 live window / round id 判定 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_current_round_support.go` 与 `awd_round_window_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_round_support.go` 不再混载两类 round support helper
- 当前轮次解析/物化与 live window / round id 判定拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService` 对外 round support 行为
- 仅改善 `contest` AWD command support 文件边界，为后续继续收口 round support 链路留下更清晰结构
