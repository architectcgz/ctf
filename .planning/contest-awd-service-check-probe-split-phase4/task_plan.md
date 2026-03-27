# Task Plan

## Goal

继续收口 `contest` AWD service check probe，把 `application/jobs/awd_service_check_probe_result.go` 从单文件拆成探测聚合 support 与结果编排两段，保持 `AWDRoundUpdater` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd service check probe 职责 | completed | 已确认单文件同时承载实例探测聚合与最终状态/错误归纳 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_service_check_probe_result.go` 与 `awd_service_check_probe_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_service_check_probe_result.go` 不再混载探测聚合细节
- 探测结果聚合与状态归纳下沉到 `awd_service_check_probe_support.go`
- `AWDRoundUpdater` 对外 service check 结果结构与行为保持不变
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外接口与 service check 行为
- 仅改善 AWD jobs probe 文件边界，为后续继续收口 service check 主流程留下更清晰结构
