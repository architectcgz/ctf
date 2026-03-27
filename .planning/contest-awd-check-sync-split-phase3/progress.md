# Progress

## 2026-03-27

- 启动 `contest-awd-check-sync-split-phase3`，目标是继续拆 `contest` AWD check sync 主流程文件。
- 盘点确认 `application/jobs/awd_check_sync.go` 同时承载三类职责：
  - service check 入口
  - service check 编排
  - 持久化与缓存写回
- 已完成文件拆分：
  - `awd_check_sync.go` 保留 scheduler/manual 入口
  - `awd_check_run.go` 承载 service check 编排
  - `awd_check_writeback.go` 承载持久化与缓存写回
- Focused 验证已完成：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
