# Progress

## 2026-03-26

- 启动 `contest-team-query-split-phase2`，目标是继续拆 `contest` team query 主流程文件。
- 盘点确认 `application/queries/team_service.go` 同时承载三类职责：
  - team info 查询与成员装配
  - team list 查询与人数统计
  - my team 查询与响应组装
- 已完成文件拆分：
  - `team_service.go` 收缩为 TeamService 类型与构造函数
  - `team_info_query.go` 承载 team info 查询与成员用户装配 helper
  - `team_list_query.go` 承载 team list / my team 查询
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/application/queries -run 'TestTeamServiceListTeamsReturnsContestNotFound' -count=1`
