# Task Plan

## Goal

继续收口 `contest` 应用层，把 `application/commands/team_membership_commands.go` 从单文件拆成更清晰的 create/join 与成员管理命令结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team command 职责 | completed | 已确认单文件同时承载 create/join、leave、dismiss、kick 四类 team 命令 |
| 2. 拆分文件结构 | completed | 已拆为 `team_membership_commands.go`、`team_create_join_commands.go`、`team_manage_commands.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 与 team command 相关定向测试 |

## Acceptance Checks

- `team_membership_commands.go` 不再承载混杂 TeamService 命令逻辑
- create/join 与成员管理命令拆到独立文件
- `contest/...` 与相关定向测试通过

## Result

- 不改 `TeamService` 对外接口
- 仅改善 `contest` commands 文件边界，为后续继续收口 team 相关命令留下更清晰结构
