# Progress

## 2026-03-27

- 启动 `contest-participation-query-split-phase2`，目标是继续拆 `contest` participation query 主流程文件。
- 盘点确认 `application/queries/participation_query.go` 同时承载两类职责：
  - registrations / announcements admin query
  - my progress 查询与用户 team 解析
- 已完成文件拆分：
  - `participation_admin_query.go` 承载 registrations / announcements 查询
  - `participation_progress_query.go` 承载 my progress 查询与 team 解析
- Focused 验证已完成：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
