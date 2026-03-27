# Progress

## 2026-03-27

- 启动 `contest-awd-list-query-split-phase4`，目标是继续拆 `contest` AWD list query 文件。
- 盘点确认 `application/queries/awd_list_query.go` 同时承载三类职责：
  - round 列表查询
  - service 列表查询
  - attack log 列表查询
- 已完成文件拆分：
  - `awd_round_list_query.go` 承载 round 列表查询
  - `awd_service_list_query.go` 承载 service 列表查询
  - `awd_attack_log_list_query.go` 承载 attack log 列表查询
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
