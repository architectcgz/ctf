# Task Plan

## Goal

继续收口 `contest` AWD jobs 层，把 `application/jobs/awd_check_sync.go` 从单文件拆成更清晰的入口、巡检编排与写回三段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd check sync 职责 | completed | 已确认单文件同时承载入口、service check 编排与持久化/缓存写回 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_check_sync.go`、`awd_check_run.go`、`awd_check_writeback.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_check_sync.go` 不再混载编排与写回细节
- service check 编排与持久化/缓存写回拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外入口方法
- 仅改善 `contest` AWD jobs 文件边界，为后续继续收口巡检同步主流程留下更清晰结构
