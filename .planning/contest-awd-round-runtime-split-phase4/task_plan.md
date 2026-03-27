# Task Plan

## Goal

继续收口 `contest` AWD round jobs，把 `application/jobs/awd_round_runtime.go` 从单文件拆成调度入口/锁管理、轮次同步运行时与对外桥接方法三段，保持 `AWDRoundUpdater` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd round runtime 职责 | completed | 已确认单文件同时承载调度入口、轮次同步和对外桥接方法 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_round_scheduler_runtime.go` 与 `awd_round_runtime_bridge.go`，`awd_round_runtime.go` 保留轮次同步逻辑 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_round_runtime.go` 不再混载调度入口和对外桥接方法
- `UpdateRoundsAt` 与 scheduler lock 逻辑拆到独立 scheduler 文件
- `SetHTTPClient` 与 `SyncRoundServiceChecks` 拆到独立 bridge 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外接口与调度行为
- 仅改善 AWD round runtime 文件边界，为后续继续收口 AWD jobs 留下更清晰结构
