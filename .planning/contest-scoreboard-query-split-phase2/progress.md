# Progress

## 2026-03-27

- 启动 `contest-scoreboard-query-split-phase2`，目标是继续拆 `contest` scoreboard query 主流程文件。
- 盘点确认 `application/queries/scoreboard_query.go` 同时承载两类职责：
  - scoreboard list / live scoreboard 分页查询
  - team rank 查询
- 已完成文件拆分：
  - `scoreboard_list_query.go` 承载 scoreboard list / live scoreboard 查询
  - `scoreboard_rank_query.go` 承载 team rank 查询
- Focused 验证已完成：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
