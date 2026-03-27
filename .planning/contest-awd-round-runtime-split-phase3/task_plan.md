# Task Plan

## Goal

继续收口 `contest` AWD jobs 层，把 `application/jobs/awd_round_updater.go` 从单文件拆成更清晰的类型/启动入口与运行时编排两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd round updater 职责 | completed | 已确认单文件同时承载类型/构造、ticker 入口与 round runtime 编排 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_round_updater.go` 与 `awd_round_runtime.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_round_updater.go` 不再混载运行时编排逻辑
- round runtime 编排拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外类型与构造函数
- 仅改善 `contest` AWD jobs 文件边界，为后续继续收口 round updater 主流程留下更清晰结构
