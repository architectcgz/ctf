# Task Plan

## Goal

继续收口 `contest` team commands，把 `application/commands/team_manage_commands.go` 从单文件拆成 leave 与 captain management 两段，保持 `TeamService` 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team manage command 职责 | completed | 已确认单文件同时承载成员离队与队长侧 dismiss/kick 两类命令流程 |
| 2. 拆分文件结构 | completed | 已拆为 `team_leave_commands.go` 与 `team_captain_manage_commands.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `team_manage_commands.go` 不再混载 leave 与 captain management 两类命令流程
- leave、dismiss、kick 命令拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `TeamService` 对外命令接口与离队/解散/踢人行为
- 仅改善 `contest` team command 文件边界，为后续继续收口 team command 链路留下更清晰结构
