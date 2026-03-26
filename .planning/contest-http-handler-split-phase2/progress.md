# Progress

## 2026-03-26

- 启动 `contest-http-handler-split-phase2`，目标是继续拆 `contest` 总 handler 主流程文件。
- 盘点确认 `api/http/handler.go` 同时承载两类职责：
  - contest create / update / get / list 入口
  - scoreboard query / freeze / unfreeze 入口
- 已完成文件拆分：
  - `handler.go` 收缩为 Handler 类型、接口与构造函数
  - `contest_handler.go` 承载 contest CRUD/query HTTP 入口
  - `scoreboard_handler.go` 承载 scoreboard HTTP 入口
- Focused 验证已完成：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
