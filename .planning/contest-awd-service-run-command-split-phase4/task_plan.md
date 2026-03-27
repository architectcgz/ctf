# Task Plan

## Goal

继续收口 `contest` AWD service commands，把 `application/commands/awd_service_run_commands.go` 从单文件拆成手动检查执行入口与响应/服务列表装配 support 两段，保持 `AWDService` 手动检查行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd service run command 职责 | completed | 已确认单文件同时承载检查执行入口与响应/服务列表装配 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_service_run_support.go`，`awd_service_run_commands.go` 保留执行入口 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_service_run_commands.go` 不再混载执行入口与响应装配
- checker run response / service list 装配拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService` 手动执行 checker 的对外接口与响应结构
- 仅改善 AWD service run command 文件边界，为后续继续收口 AWD service commands 留下更清晰结构
