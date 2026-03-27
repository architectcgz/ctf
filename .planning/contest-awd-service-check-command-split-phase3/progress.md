# Progress

## 2026-03-27

- 启动 `contest-awd-service-check-command-split-phase3`，目标是继续拆 `contest` AWD service check command 主流程文件。
- 盘点确认 `application/commands/awd_service_check_commands.go` 同时承载两类职责：
  - run current/selected round checks
  - upsert manual service check
- 已完成文件拆分：
  - `awd_service_run_commands.go` 承载 run current/selected round checks
  - `awd_service_upsert_commands.go` 承载 upsert manual service check
- Focused 验证已完成：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
