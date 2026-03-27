# Task Plan

## Goal

继续收口 `contest` AWD jobs 层，把 `application/jobs/awd_checks.go` 从单文件拆成更清晰的结果类型定义与巡检结果生成两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd checks 职责 | completed | 已确认单文件同时承载结果类型定义与 service check 结果生成流程 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_checks.go` 与 `awd_service_check_result.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_checks.go` 不再混载结果类型定义与结果生成主流程
- service check 结果生成拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外类型与相关返回结构
- 仅改善 `contest` AWD jobs 文件边界，为后续继续收口巡检链路留下更清晰结构
