# Task Plan

## Goal

继续收口 `contest` application commands，把 `application/commands/participation_registration_commands.go` 从单文件拆成报名与审核两段，保持 `ParticipationService` 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 participation registration command 职责 | completed | 已确认单文件同时承载 contest 报名与 registration 审核两类命令流程 |
| 2. 拆分文件结构 | completed | 已拆为 `participation_register_commands.go` 与 `participation_review_commands.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `participation_registration_commands.go` 不再混载报名与审核两类命令流程
- 报名与审核命令拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ParticipationService` 对外命令接口与报名/审核行为
- 仅改善 `contest` participation command 文件边界，为后续继续收口 participation command 链路留下更清晰结构
