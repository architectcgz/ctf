# Task Plan

## Goal

继续收口 `contest` AWD 后台任务，把 `application/jobs/awd_checks.go` 从单文件拆成更清晰的巡检编排与辅助结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD service check 职责 | completed | 已确认 `awd_checks.go` 同时承载类型定义、巡检编排、live cache 判定与实例装载辅助 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_checks.go`、`awd_check_sync.go`、`awd_check_support.go` |
| 3. focused 验证 | completed | 已运行 `./internal/module/contest/...` 定向测试 |

## Acceptance Checks

- `awd_checks.go` 仅保留巡检结果模型与结果聚合逻辑
- service check 编排流程与 support helper 拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 AWD scheduler / 手动巡检行为
- 仅改善 jobs 层文件边界，为后续继续拆 AWD updater 留出更清晰落点
