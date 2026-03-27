# Task Plan

## Goal

继续收口 `contest` AWD jobs 层，把 `application/jobs/awd_probe.go` 从单文件拆成更清晰的 probe 结果类型定义与 probe runtime 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd probe 职责 | completed | 已确认单文件同时承载 probe 结果类型定义与 probe runtime 逻辑 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_probe.go` 与 `awd_probe_runtime.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_probe.go` 不再混载 probe runtime 逻辑
- probe runtime 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外行为与 probe 结果结构
- 仅改善 `contest` AWD jobs 文件边界，为后续继续收口 probe 主流程留下更清晰结构
