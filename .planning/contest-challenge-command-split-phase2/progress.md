# Progress

## 2026-03-26

- 启动 `contest-challenge-command-split-phase2`，目标是继续拆 `contest` challenge command 主流程文件。
- 盘点确认 `application/commands/challenge_service.go` 同时承载三类职责：
  - add challenge to contest
  - remove challenge from contest
  - update challenge in contest
- 已完成文件拆分：
  - `challenge_service.go` 收缩为 ChallengeService 类型与构造函数
  - `challenge_add_commands.go` 承载 add challenge 命令
  - `challenge_manage_commands.go` 承载 remove / update 命令
  - `challenge_support.go` 承载 contest 可变性校验 helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
