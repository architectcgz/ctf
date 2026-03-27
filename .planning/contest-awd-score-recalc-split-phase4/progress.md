# Progress

## 2026-03-27

- 启动 `contest-awd-score-recalc-split-phase4`，目标是继续拆 `contest` AWD score recalc 文件。
- 盘点确认 `infrastructure/awd_score_recalc.go` 同时承载三类职责：
  - contest/team/service/attack 数据读取
  - defense/attack 总分聚合
  - team 总分与 `last_solve_at` 写回
- 已完成文件拆分：
  - `awd_score_recalc.go` 保留重算入口编排
  - `awd_score_recalc_loader.go` 承载 team/service/attack 读取
  - `awd_score_recalc_writeback.go` 承载 team 分数与 `last_solve_at` 写回
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
