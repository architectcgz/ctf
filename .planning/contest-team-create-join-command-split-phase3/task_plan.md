# Task Plan

## Goal

继续收口 `contest` team commands，把 `application/commands/team_create_join_commands.go` 从单文件拆成 create 与 join 两段，保持 `TeamService` 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team create/join command 职责 | completed | 已确认单文件同时承载 create team 与 join team 两类命令流程 |
| 2. 拆分文件结构 | completed | 已拆为 `team_create_commands.go` 与 `team_join_commands.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `team_create_join_commands.go` 不再混载 create 与 join 两类命令流程
- create team 与 join team 命令拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `TeamService` 对外命令接口与建队/入队行为
- 仅改善 `contest` team command 文件边界，为后续继续收口 team command 链路留下更清晰结构
