# Progress

## 2026-03-27

- 启动 `contest-awd-query-split-phase3`，目标是继续拆 `contest` AWD query 主流程文件。
- 盘点确认 `application/queries/awd_query.go` 同时承载两类职责：
  - round / services / attack logs list 查询
  - round summary 查询
- 已完成文件拆分：
  - `awd_list_query.go` 承载 round / services / attack logs list 查询
  - `awd_summary_query.go` 承载 round summary 查询
- Focused 验证已完成：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
