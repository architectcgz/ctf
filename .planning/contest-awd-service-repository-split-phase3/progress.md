# Progress

## 2026-03-27

- 启动 `contest-awd-service-repository-split-phase3`，目标是继续拆 `contest` AWD infrastructure repository 文件。
- 盘点确认 `infrastructure/awd_service_repository.go` 同时承载三类职责：
  - service instance 查询
  - team service 持久化
  - attack log / victim impact 持久化
- 已完成文件拆分：
  - `awd_service_instance_repository.go` 承载 service instance 查询
  - `awd_team_service_repository.go` 承载 team service 持久化
  - `awd_attack_log_repository.go` 承载 attack log / victim impact 持久化
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
