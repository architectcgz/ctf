# Progress

## 2026-03-27

- 启动 `contest-awd-relation-repository-split-phase4`，目标是继续拆 `contest` AWD relation repository 文件。
- 盘点确认 `infrastructure/awd_relation_repository.go` 同时承载两类职责：
  - contest / challenge relation
  - team / member / registration relation
- 已完成文件拆分：
  - `awd_contest_relation_repository.go` 承载 contest / challenge relation
  - `awd_team_relation_repository.go` 承载 team / member / registration relation
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
