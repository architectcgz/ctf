# Task Plan

## Goal

继续收口 `contest` AWD command 层，把 `application/commands/awd_service_check_commands.go` 从单文件拆成更清晰的 run 与 upsert 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd service check command 职责 | completed | 已确认单文件同时承载 run checks 与 upsert service check 两类命令 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_service_run_commands.go` 与 `awd_service_upsert_commands.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_service_check_commands.go` 不再混载 run 与 upsert 两类命令
- run checks 与 upsert service check 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService` 对外命令接口
- 仅改善 `contest` AWD command 文件边界，为后续继续收口 AWD command 主流程留下更清晰结构
