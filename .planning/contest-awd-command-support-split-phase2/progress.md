# Progress

## 2026-03-26

- 启动 `contest-awd-command-support-split-phase2`，目标是继续拆 `contest` AWD 写侧 helper 文件。
- 盘点确认 `application/commands/awd_support.go` 同时承载三类职责：
  - AWD contest/round/challenge 校验与 team/challenge 装载
  - 当前轮次解析、materialize 与 live window 判定
  - round flag 解析与 previous-round grace period helper
- 已完成文件拆分：
  - `awd_support.go` 收缩为 package 占位
  - `awd_validation_support.go` 承载校验与 team/challenge 装载
  - `awd_round_support.go` 承载当前轮次解析与 materialize helper
  - `awd_flag_support.go` 承载 round flag 与 grace period helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
