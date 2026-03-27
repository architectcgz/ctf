# Progress

## 2026-03-27

- 启动 `contest-scoreboard-list-query-split-phase3`，目标是继续拆 `contest` scoreboard list query 文件。
- 盘点确认 `application/queries/scoreboard_list_query.go` 同时承载两类职责：
  - scoreboard 查询入口
  - 分页 / redis key / item 组装 support
- 已完成文件拆分：
  - `scoreboard_list_query.go` 承载 scoreboard 查询主流程
  - `scoreboard_list_support.go` 承载分页 / redis key / item 组装 support
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
