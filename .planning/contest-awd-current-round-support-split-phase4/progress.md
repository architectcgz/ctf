# Progress

## 2026-03-27

- 启动 `contest-awd-current-round-support-split-phase4`，目标是继续拆 `contest` AWD current round support 文件。
- 盘点确认 `application/commands/awd_current_round_support.go` 同时承载三类职责：
  - 当前 active round 的主路径解析
  - active round materialize 与二次读取
  - running round / redis current-round fallback
- 已完成文件拆分：
  - `awd_current_round_active_support.go` 承载 active round materialize 主路径
  - `awd_current_round_fallback_support.go` 承载 running round / redis fallback
  - `awd_current_round_support.go` 仅保留入口编排
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
