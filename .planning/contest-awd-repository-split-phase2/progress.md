# Progress

## 2026-03-26

- 启动 `contest-awd-repository-split-phase2`，目标是继续拆 `contest` AWD infrastructure 文件。
- 盘点确认 `infrastructure/awd_repository.go` 同时承载四类职责：
  - `AWDRepository` 基础仓储定义与事务 helper
  - round 相关持久化
  - contest/team/challenge 关系查询
  - service check 与 attack log 持久化
- 已完成文件拆分：
  - `awd_repository.go` 保留基础仓储定义与共享 helper
  - `awd_round_repository.go` 承载 round 持久化
  - `awd_relation_repository.go` 承载 contest/team/challenge 关系查询
  - `awd_service_repository.go` 承载 service check 与 attack log 持久化
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
