# Task Plan

## Goal

继续收口 `contest` team create command，把 `application/commands/team_create_commands.go` 从单文件拆成入口编排与邀请码重试创建 support 两段，保持 `TeamService.CreateTeam` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team create command 职责 | completed | 已确认单文件同时承载前置校验与邀请码重试创建 |
| 2. 拆分文件结构 | completed | 已拆为 `team_create_retry_support.go`，`team_create_commands.go` 保留入口编排 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `team_create_commands.go` 不再混载邀请码重试创建细节
- 邀请码重试创建与冲突错误映射拆到独立 support 文件
- `TeamService.CreateTeam` 对外返回与错误语义保持不变
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `TeamService.CreateTeam` 对外接口与业务行为
- 仅改善 team create command 文件边界，为后续继续收口 team commands 留下更清晰结构
