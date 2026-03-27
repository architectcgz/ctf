# Progress

## 2026-03-27

- 启动 `contest-awd-summary-query-split-phase3`，目标是继续拆 `contest` AWD summary query 文件。
- 盘点确认 `application/queries/awd_summary_query.go` 同时承载两类职责：
  - round summary 查询入口
  - metrics / items 聚合与排序
- 已完成文件拆分：
  - `awd_summary_query.go` 承载 round summary 查询主流程
  - `awd_summary_support.go` 承载 metrics / items 聚合与排序 support
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
