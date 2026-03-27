# Task Plan

## Goal

继续收口 `contest` AWD summary queries，把 `application/queries/awd_summary_support.go` 从单文件拆成主汇总入口、service 维度汇总与 attack 维度汇总三段，保持 AWD round summary 计算行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd summary support 职责 | completed | 已确认单文件同时承载主汇总入口、service 维度汇总与 attack 维度汇总 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_summary_service_support.go` 与 `awd_summary_attack_support.go`，`awd_summary_support.go` 保留主汇总入口 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_summary_support.go` 不再混载 service 与 attack 两段汇总细节
- service 维度汇总拆到独立 support 文件
- attack 维度汇总拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 AWD round summary 的输出结构、排序与计数逻辑
- 仅改善 AWD summary support 文件边界，为后续继续收口 AWD queries 留下更清晰结构
