# Progress

## 2026-03-27

- 启动 `contest-team-manage-command-split-phase3`，目标是继续拆 `contest` team manage command 文件。
- 盘点确认 `application/commands/team_manage_commands.go` 同时承载两类职责：
  - 成员 leave team
  - captain 侧 dismiss team / kick member
- 已完成文件拆分：
  - `team_leave_commands.go` 承载成员离队命令
  - `team_captain_manage_commands.go` 承载队长侧 dismiss / kick 命令
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
