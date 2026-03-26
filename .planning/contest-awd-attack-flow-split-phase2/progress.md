# Progress

## 2026-03-26

- 启动 `contest-awd-attack-flow-split-phase2`，目标是继续拆 `contest` AWD attack 主流程文件。
- 盘点确认 `application/commands/awd_attack_commands.go` 同时承载两类职责：
  - 手工 `CreateAttackLog` / `createAttackLog` 流程
  - `SubmitAttack` flag 提交流程
- 已完成文件拆分：
  - `awd_attack_commands.go` 收缩为 package 占位
  - `awd_attack_log_commands.go` 承载手工 attack log 流程
  - `awd_attack_submit_commands.go` 承载 flag 提交流程
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
