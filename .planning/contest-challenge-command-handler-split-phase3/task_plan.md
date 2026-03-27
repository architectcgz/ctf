# Task Plan

## Goal

继续收口 `contest` challenge HTTP handlers，把 `api/http/challenge_command_handler.go` 从单文件拆成 add 与 manage 两段，保持 `ChallengeHandler` 对外路由与响应行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 challenge command handler 职责 | completed | 已确认单文件同时承载 add challenge 与 remove/update 两类 HTTP 命令入口 |
| 2. 拆分文件结构 | completed | 已拆为 `challenge_add_handler.go` 与 `challenge_manage_handler.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `challenge_command_handler.go` 不再混载 add challenge 与 remove/update 两类 HTTP 命令入口
- add 与 manage handler 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ChallengeHandler` 对外路由、参数校验与响应行为
- 仅改善 `contest` challenge HTTP 文件边界，为后续继续收口 challenge handler 链路留下更清晰结构
