# Task Plan

## Goal

继续收口 `contest` submission helper，把 `submission_support.go` 从单文件拆成更清晰的 score 与 error support 结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 submission helper 职责 | completed | 已确认 `submission_support.go` 同时承载动态计分、score update 组装、错误映射与唯一约束判断 |
| 2. 拆分文件结构 | completed | 已拆为 `submission_support.go`、`submission_score_support.go`、`submission_error_support.go` |
| 3. focused 验证 | completed | 已运行 `./internal/module/contest/...` 定向测试 |

## Acceptance Checks

- `submission_support.go` 不再承载混杂 helper 实现
- score helper 与 error helper 拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 submission commands 对外接口
- 仅改善 submission helper 边界，为后续继续收 submission 写侧留出更清晰落点
