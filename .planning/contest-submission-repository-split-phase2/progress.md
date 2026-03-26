# Progress

## 2026-03-26

- 启动 `contest-submission-repository-split-phase2`，目标是继续拆 `contest` submission repository 主流程文件。
- 盘点确认 `infrastructure/submission_repository.go` 同时承载多类职责：
  - 事务包装
  - registration / challenge lookup
  - submission 写入与分数更新
  - challenge lock、首血更新与 team score 变更
- 已完成文件拆分：
  - `submission_repository.go` 保留 Repository 类型、构造函数、`WithDB`、`dbWithContext` 与事务包装
  - `submission_lookup_repository.go` 承载 registration / challenge lookup
  - `submission_write_repository.go` 承载 submission 写入与 submission score 更新
  - `submission_score_repository.go` 承载计分相关锁、统计与 team score 更新
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
