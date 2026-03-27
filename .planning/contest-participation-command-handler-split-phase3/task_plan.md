# Task Plan

## Goal

继续收口 `contest` participation HTTP handlers，把 `api/http/participation_command_handler.go` 从单文件拆成 registration/review 与 announcement 两段，保持 `ParticipationHandler` 对外路由与响应行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 participation command handler 职责 | completed | 已确认单文件同时承载 registration/review 与 announcement 两类 HTTP 命令入口 |
| 2. 拆分文件结构 | completed | 已拆为 `participation_registration_handler.go` 与 `participation_announcement_handler.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `participation_command_handler.go` 不再混载 registration/review 与 announcement 两类 HTTP 命令入口
- registration/review 与 announcement handler 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `ParticipationHandler` 对外路由、参数校验与响应行为
- 仅改善 `contest` participation HTTP 文件边界，为后续继续收口 participation handler 链路留下更清晰结构
