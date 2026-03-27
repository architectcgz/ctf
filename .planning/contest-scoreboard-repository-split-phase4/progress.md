# Progress

## 2026-03-27

- 启动 `contest-scoreboard-repository-split-phase4`，目标是继续拆 `contest` scoreboard repository 文件。
- 盘点确认 `infrastructure/contest_scoreboard_repository.go` 同时承载三类职责：
  - `FindScoreboardTeamStats` 公共入口与结果映射
  - AWD / 非 AWD scoreboard stats 查询分支
  - aggregate time parsing support
- 已完成文件拆分：
  - `contest_scoreboard_mode_repository.go` 承载 AWD / 非 AWD 查询分支
  - `contest_scoreboard_time_support.go` 承载 aggregate time parsing support
  - `contest_scoreboard_repository.go` 仅保留公共入口与结果映射
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
