# Task Plan

## Goal

继续收口 `contest` submission 写侧，把 `submission_submit.go` 从单文件拆成更清晰的提交流程、team 解析和计分处理结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 submission 提交流程职责 | completed | 已确认 `submission_submit.go` 同时承载入口流程、team 解析与正确提交后的计分事务 |
| 2. 拆分文件结构 | completed | 已拆为 `submission_submit.go`、`submission_validation.go`、`submission_scoring.go` |
| 3. focused 验证 | completed | 已运行 `./internal/module/contest/...` 定向测试 |

## Acceptance Checks

- `submission_submit.go` 仅保留提交入口流程
- team 解析与计分事务逻辑拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 submission commands 对外接口
- 仅改善 commands 层文件边界，为后续继续拆 submission 写侧细节留出更清晰落点
