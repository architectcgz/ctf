# Task Plan

## Goal

继续收口 `contest` infrastructure，把 `infrastructure/submission_repository.go` 从单文件拆成更清晰的 lookup、写入与计分更新结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 submission repository 职责 | completed | 已确认单文件同时承载事务包装、lookup、submission 写入与计分更新 |
| 2. 拆分文件结构 | completed | 已拆为 `submission_repository.go`、`submission_lookup_repository.go`、`submission_write_repository.go`、`submission_score_repository.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 定向测试 |

## Acceptance Checks

- `submission_repository.go` 不再承载混杂 SubmissionRepository 逻辑
- lookup / submission 写入 / 计分更新拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 `SubmissionRepository` 对外接口
- 仅改善 `contest` infrastructure 文件边界，为后续继续收口 submission 仓储留下更清晰结构
