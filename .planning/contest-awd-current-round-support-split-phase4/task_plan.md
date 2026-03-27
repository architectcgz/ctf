# Task Plan

## Goal

继续收口 `contest` AWD command support，把 `application/commands/awd_current_round_support.go` 从单文件拆成入口编排、active round materialize 主路径与 running/redis fallback 三段，保持 `AWDService` 当前轮次解析行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd current round support 职责 | completed | 已确认单文件同时承载 active round 主路径、materialize 与 running/redis fallback |
| 2. 拆分文件结构 | completed | 已拆为 `awd_current_round_active_support.go` 与 `awd_current_round_fallback_support.go`，`awd_current_round_support.go` 保留入口编排 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_current_round_support.go` 不再混载 active round 主路径与 fallback 逻辑
- active round materialize 主路径拆到独立 support 文件
- running round / redis fallback 拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService` 当前轮次解析 helper 的对外函数签名与行为
- 仅改善 AWD current round support 文件边界，为后续继续收口 AWD command support 留下更清晰结构
