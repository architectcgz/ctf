# Task Plan

## Goal

继续收口 `contest` AWD jobs，把 `application/jobs/awd_service_check_result.go` 从单文件结果组装器拆成无实例 fallback 与多实例 probe 聚合两段，保持 `AWDRoundUpdater` 服务检查行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd service check result 职责 | completed | 已确认单文件同时承载 no-instance fallback 与 probe 聚合序列化 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_service_check_empty_result.go` 与 `awd_service_check_probe_result.go`，`awd_service_check_result.go` 保留编排 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_service_check_result.go` 不再混载空实例 fallback 与 probe 聚合
- 无实例 fallback 结果拆到独立文件
- 多实例 probe 聚合与序列化拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外服务检查流程与结果结构
- 仅改善 AWD jobs 中文件边界，为后续继续收口 AWD runtime/check 链路留下更清晰结构
