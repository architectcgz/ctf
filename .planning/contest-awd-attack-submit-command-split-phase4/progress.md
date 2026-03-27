# Progress

## 2026-03-27

- 启动 `contest-awd-attack-submit-command-split-phase4`，目标是继续拆 `contest` AWD attack submit command 文件。
- 盘点确认 `application/commands/awd_attack_submit_commands.go` 同时承载三类职责：
  - SubmitAttack 入口编排
  - 提交上下文解析（contest/round/challenge/team/accepted flags）
  - 提交 flag 命中判定
- 已完成文件拆分：
  - `awd_attack_submit_support.go` 承载提交上下文解析与命中判定
  - `awd_attack_submit_commands.go` 保留入口编排
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
