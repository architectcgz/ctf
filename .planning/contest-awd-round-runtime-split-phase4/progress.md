# Progress

## 2026-03-27

- 启动 `contest-awd-round-runtime-split-phase4`，目标是继续拆 `contest` AWD round runtime 文件。
- 盘点确认 `application/jobs/awd_round_runtime.go` 同时承载三类职责：
  - scheduler 入口与分布式锁
  - contest round 同步运行时
  - 对外桥接方法（HTTP client / 手动同步）
- 已完成文件拆分：
  - `awd_round_scheduler_runtime.go` 承载 scheduler 入口与锁管理
  - `awd_round_runtime_bridge.go` 承载对外桥接方法
  - `awd_round_runtime.go` 保留 contest round 同步运行时
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
