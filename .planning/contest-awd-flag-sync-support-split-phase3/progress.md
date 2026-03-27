# Progress

## 2026-03-27

- 启动 `contest-awd-flag-sync-support-split-phase3`，目标是继续拆 `contest` AWD flag sync 主流程文件。
- 盘点确认 `application/jobs/awd_round_flag_sync.go` 同时承载两类职责：
  - flag sync 主流程
  - round/assignment/support helper
- 已完成文件拆分：
  - `awd_round_flag_sync.go` 保留 flag sync 主流程
  - `awd_round_flag_support.go` 承载 round 查询、assignment 构造与 TTL helper
- Focused 验证已完成：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
