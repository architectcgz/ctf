# Progress

## 2026-03-26

- 启动 `contest-awd-query-split-phase2`，目标是继续拆 `contest` AWD 读侧查询文件。
- 盘点确认 `application/queries/awd_service.go` 同时承载三类职责：
  - `AWDService` 构造与依赖持有
  - `ListRounds / ListServices / ListAttackLogs / GetRoundSummary` 查询流程
  - `ensureAWDContest / ensureAWDRound / loadContestTeams` 校验与装载辅助
- 已完成文件拆分：
  - `awd_service.go` 保留 service 定义与构造器
  - `awd_query.go` 承载 AWD 读侧查询流程
  - `awd_support.go` 承载 contest/round 校验与 team 装载 helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
