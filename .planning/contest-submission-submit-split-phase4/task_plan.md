# Task Plan

## Goal

继续收口 `contest` submission command，把 `application/commands/submission_submit.go` 从单文件入口拆成前置校验、错误提交落库与入口编排三段，保持 `SubmissionService.SubmitFlagInContest` 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 submission submit 职责 | completed | 已确认单文件同时承载前置校验、错误提交分支与正确提交流程编排 |
| 2. 拆分文件结构 | completed | 已拆为 `submission_submit_validation.go` 与 `submission_incorrect_submit.go`，`submission_submit.go` 保留入口编排 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `submission_submit.go` 不再混载前置校验与错误提交落库
- contest submission 前置校验拆到独立 validation 文件
- incorrect submission 处理拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `SubmissionService.SubmitFlagInContest` 对外接口与提交行为
- 仅改善 submission submit 链路文件边界，为后续继续收口 submission application command 留下更清晰结构
