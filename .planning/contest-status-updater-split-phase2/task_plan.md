# Task Plan

## Goal

继续收口 `contest` jobs 层，把 `application/jobs/status_updater.go` 从单文件拆成更清晰的 runner 与 support 两段，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 status updater 职责 | completed | 已确认单文件同时承载 ticker 入口、状态迁移主流程与 snapshot/cleanup helper |
| 2. 拆分文件结构 | completed | 已拆为 `status_updater.go`、`status_update_runner.go`、`status_update_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `status_updater.go` 不再混载 runner 与 support 具体逻辑
- 状态迁移 runner 与 snapshot/cleanup helper 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `StatusUpdater` 对外类型与构造函数
- 仅改善 `contest` jobs 文件边界，为后续继续收口后台任务主流程留下更清晰结构
