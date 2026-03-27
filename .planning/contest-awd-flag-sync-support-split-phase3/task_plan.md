# Task Plan

## Goal

继续收口 `contest` AWD jobs 层，把 `application/jobs/awd_round_flag_sync.go` 从单文件拆成更清晰的 flag sync 主流程与 round/assignment support 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd round flag sync 职责 | completed | 已确认单文件同时承载 flag sync 主流程与 round/assignment support helper |
| 2. 拆分文件结构 | completed | 已拆为 `awd_round_flag_sync.go` 与 `awd_round_flag_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_round_flag_sync.go` 不再混载 support helper
- round 查询、assignment 构造与 TTL helper 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外 flag sync 行为
- 仅改善 `contest` AWD jobs 文件边界，为后续继续收口 flag sync 主流程留下更清晰结构
