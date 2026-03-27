# Progress

## 2026-03-27

- 启动 `contest-awd-summary-support-split-phase4`，目标是继续拆 `contest` AWD summary support 文件。
- 盘点确认 `application/queries/awd_summary_support.go` 同时承载三类职责：
  - AWD round summary 主汇总入口
  - service 维度 metrics 与 team summary 累加
  - attack 维度 metrics 与 team summary 累加
- 已完成文件拆分：
  - `awd_summary_service_support.go` 承载 service 维度汇总
  - `awd_summary_attack_support.go` 承载 attack 维度汇总
  - `awd_summary_support.go` 保留主汇总入口
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
