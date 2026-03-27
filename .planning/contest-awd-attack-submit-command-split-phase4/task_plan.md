# Task Plan

## Goal

继续收口 `contest` AWD attack submit command，把 `application/commands/awd_attack_submit_commands.go` 从单文件拆成入口编排与提交上下文解析/命中判定 support 两段，保持 `AWDService.SubmitAttack` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd attack submit command 职责 | completed | 已确认单文件同时承载提交流程编排、上下文解析与 flag 命中判定 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_attack_submit_support.go`，`awd_attack_submit_commands.go` 保留入口编排 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_attack_submit_commands.go` 不再混载上下文解析与 flag 命中判定
- 提交上下文解析拆到独立 support 文件
- flag 命中判定拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService.SubmitAttack` 对外接口、参数语义与提交流程行为
- 仅改善 AWD attack submit command 文件边界，为后续继续收口 AWD commands 留下更清晰结构
