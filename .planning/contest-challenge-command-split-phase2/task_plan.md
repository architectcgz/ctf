# Task Plan

## Goal

继续收口 `contest` 应用层，把 `application/commands/challenge_service.go` 从单文件拆成更清晰的 add 与 manage 命令结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 challenge command 职责 | completed | 已确认单文件同时承载 add、remove、update 三类 challenge 命令流程 |
| 2. 拆分文件结构 | completed | 已拆为 `challenge_service.go`、`challenge_add_commands.go`、`challenge_manage_commands.go`、`challenge_support.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 定向测试 |

## Acceptance Checks

- `challenge_service.go` 不再承载混杂 ChallengeService 命令逻辑
- add / remove / update 命令拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 `ChallengeService` 对外接口
- 仅改善 `contest` commands 文件边界，为后续继续收口 challenge 相关命令留下更清晰结构
