# Task Plan

## Goal

继续收口 `contest` team HTTP handlers，把 `api/http/team_command_handler.go` 从单文件拆成 create/join 与 leave/dismiss/kick 两段，保持 `TeamHandler` 对外路由行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team command handler 职责 | completed | 已确认单文件同时承载 create/join 与 leave/dismiss/kick 两类 HTTP 命令入口 |
| 2. 拆分文件结构 | completed | 已拆为 `team_create_join_handler.go` 与 `team_manage_handler.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `team_command_handler.go` 不再混载 create/join 与 leave/dismiss/kick 两类 HTTP 命令入口
- create/join 与 leave/dismiss/kick handler 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `TeamHandler` 对外路由与响应行为
- 仅改善 `contest` team HTTP 文件边界，为后续继续收口 team handler 链路留下更清晰结构
