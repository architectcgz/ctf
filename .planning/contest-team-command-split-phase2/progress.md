# Progress

## 2026-03-26

- 启动 `contest-team-command-split-phase2`，目标是继续拆 `contest` team command 主流程文件。
- 盘点确认 `application/commands/team_membership_commands.go` 同时承载多类职责：
  - create team
  - join team
  - leave / dismiss team
  - kick member
- 已完成文件拆分：
  - `team_membership_commands.go` 收缩为 package 占位
  - `team_create_join_commands.go` 承载 create / join 命令
  - `team_manage_commands.go` 承载 leave / dismiss / kick 命令
  - `team_support.go` 补充 contest team 状态校验 helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/application/commands -run 'TestTeamServiceCreateTeamRequiresApprovedRegistration' -count=1`
