# Progress

## 2026-03-27

- 启动 `contest-participation-admin-query-split-phase4`，目标是继续拆 `contest` participation admin query 文件。
- 盘点确认 `application/queries/participation_admin_query.go` 同时承载两类职责：
  - registration admin list query
  - announcement list query
- 已完成文件拆分：
  - `participation_registration_admin_query.go` 承载 registration admin list query
  - `participation_announcement_query.go` 承载 announcement list query
  - `participation_query_support.go` 承载共享 contest existence 校验
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
