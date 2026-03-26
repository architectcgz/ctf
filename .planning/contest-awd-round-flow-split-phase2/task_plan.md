# Task Plan

## Goal

继续收口 `contest` AWD 主流程，把 `awd_round_commands.go` 从单文件拆成更清晰的 round 管理与 service check 命令结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD round command 职责 | completed | 已确认 `awd_round_commands.go` 同时承载 round 管理、手动巡检触发、服务状态上报 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_round_commands.go`、`awd_round_admin_commands.go`、`awd_service_check_commands.go` |
| 3. focused 验证 | completed | 已运行 `./internal/module/contest/...` 定向测试 |

## Acceptance Checks

- `awd_round_commands.go` 不再承载混杂命令流程
- round 管理与 service check 命令拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 AWD commands 对外接口
- 仅改善 AWD 主流程文件边界，为后续继续拆 `attack` 侧主流程留出更清晰落点
