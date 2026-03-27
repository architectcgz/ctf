# Progress

## 2026-03-27

- 启动 `contest-awd-score-recalc-split-phase3`，目标是继续拆 `contest` AWD 积分重算文件。
- 盘点确认 `infrastructure/awd_score_recalc.go` 同时承载两类职责：
  - AWD 积分重算 / cache 同步入口
  - defense / attack 积分归并与 official totals 判定
- 已完成文件拆分：
  - `awd_score_recalc.go` 承载重算 / cache 同步主流程
  - `awd_score_recalc_support.go` 承载 defense / attack 积分归并与 official totals 判定
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
