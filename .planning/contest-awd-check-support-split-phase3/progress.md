# Progress

## 2026-03-27

- 启动 `contest-awd-check-support-split-phase3`，目标是继续拆 `contest` AWD check support 文件。
- 盘点确认 `application/jobs/awd_check_support.go` 同时承载两类职责：
  - live service status cache 判定
  - contest service instance 装载
- 已完成文件拆分：
  - `awd_check_cache_support.go` 承载 live cache 判定
  - `awd_check_instance_support.go` 承载 service instance 装载
- Focused 验证已完成：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
