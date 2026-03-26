# Findings

- `contest` 已完成 `commands / queries / jobs / ports / domain / infrastructure` 物理分层，但应用层多个服务仍共享过宽的 `ports.Repository`。
- `contest/application/commands/contest_service.go` 只需要 `Create / FindByID / Update`，却依赖包含列表、状态调度、计分板聚合等能力的宽接口。
- `contest/application/queries/scoreboard_service.go` 只需要 `FindByID / FindTeamsByIDs / FindScoreboardTeamStats`，当前构造器却绑定全量 `ports.Repository`。
- `contest/application/jobs/status_updater.go` 只需要 `ListByStatusesAndTimeRange / UpdateStatus`，测试桩却被迫实现一整套无关方法。
- 第一刀优先收紧端口比继续拆 composition 更有价值；先让应用层按用例依赖最小接口，后续再继续拆装配会更稳。
- 当前阶段已完成：
  - 应用层不再依赖 legacy 宽 `contestports.Repository`
  - composition 已不再持有 concrete contest repo 字段
  - `contest` phase2 的端口与装配收口目标已达成
