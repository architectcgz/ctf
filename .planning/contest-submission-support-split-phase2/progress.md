# Progress

## 2026-03-26

- 启动 `contest-submission-support-split-phase2`，目标是继续拆 `contest` submission helper 文件。
- 盘点确认 `application/commands/submission_support.go` 同时承载两类职责：
  - 动态计分与 score update 组装 helper
  - 错误映射与唯一约束判断 helper
- 已完成文件拆分：
  - `submission_support.go` 收缩为 package 占位
  - `submission_score_support.go` 承载动态计分与 score update helper
  - `submission_error_support.go` 承载错误映射与唯一约束判断 helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
