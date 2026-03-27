# Task Plan

## Goal

继续收口 `contest` submission application command，把 `application/commands/submission_scoring.go` 从单文件入口拆成事务内计分更新与事务外榜单同步两段，保持 `SubmissionService` 提交流程与外部行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 submission scoring 职责 | completed | 已确认单文件同时承载事务内正确提交计分与事务外 scoreboard sync |
| 2. 拆分文件结构 | completed | 已拆为 `submission_score_transaction.go` 与 `submission_scoreboard_sync.go`，`submission_scoring.go` 保留入口编排 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `submission_scoring.go` 不再混载事务内计分写入与事务外 scoreboard sync
- 正确提交计分逻辑拆到独立 transaction 文件
- scoreboard 增量同步与失败回退 rebuild 拆到独立 sync 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `SubmissionService` 对外接口与正确提交记分行为
- 仅改善 submission command 文件边界，为后续继续收口 submission / scoreboard 链路留下更清晰结构
