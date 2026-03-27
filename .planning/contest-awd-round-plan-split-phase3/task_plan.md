# Task Plan

## Goal

继续收口 `contest` AWD jobs，把 `application/jobs/awd_round_plan.go` 从单文件拆成 round plan 计算、round reconcile、round lock 三段，保持 `AWDRoundUpdater` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd round plan 职责 | completed | 已确认单文件同时承载 round plan 计算、round reconcile 落库、redis lock |
| 2. 拆分文件结构 | completed | 已保留 `awd_round_plan.go` 承载 plan 计算，并新增 `awd_round_reconcile.go`、`awd_round_lock.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_round_plan.go` 不再混载 plan 计算、round reconcile、redis lock 三类职责
- round reconcile 与 round lock 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` round runtime 对外行为
- 仅改善 `contest` AWD jobs 文件边界，为后续继续收口 round runtime 链路留下更清晰结构
