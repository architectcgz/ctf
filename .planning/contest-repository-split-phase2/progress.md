# Progress

## 2026-03-26

- 启动 `contest-repository-split-phase2`，目标是继续拆 `contest` 通用 repository 主流程文件。
- 盘点确认 `infrastructure/repository.go` 同时承载多类职责：
  - contest 创建、更新、读取
  - 状态调度查询与状态更新
  - team lookup 查询
  - scoreboard stats 聚合与时间解析 helper
- 已完成文件拆分：
  - `repository.go` 收缩为 Repository 类型与构造函数
  - `contest_repository.go` 承载 contest CRUD 与状态调度查询
  - `contest_team_lookup_repository.go` 承载 team lookup 查询
  - `contest_scoreboard_repository.go` 承载 scoreboard stats 聚合与时间解析 helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
