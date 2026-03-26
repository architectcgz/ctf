# Progress

## 2026-03-26

- 启动 `contest-awd-check-split-phase2`，目标是继续拆 `contest` AWD 后台巡检文件。
- 盘点确认 `application/jobs/awd_checks.go` 同时承载三类职责：
  - 巡检结果模型与 JSON 序列化
  - `syncRoundServiceChecks / runRoundServiceChecks` 巡检编排
  - live status cache 判定与实例装载 helper
- 已完成文件拆分：
  - `awd_checks.go` 保留巡检结果模型与 `checkTeamChallengeServices`
  - `awd_check_sync.go` 承载轮次 service check 编排
  - `awd_check_support.go` 承载 live cache 判定与实例装载 helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
