# Progress

## 2026-03-27

- 启动 `contest-submission-scoring-split-phase4`，目标是继续拆 `contest` submission scoring 文件。
- 盘点确认 `application/commands/submission_scoring.go` 同时承载两类职责：
  - 事务内正确提交计分、首血更新与 team score 累加
  - 事务外 scoreboard 增量同步与失败时 rebuild 回退
- 已完成文件拆分：
  - `submission_score_transaction.go` 承载事务内正确提交计分流程
  - `submission_scoreboard_sync.go` 承载事务外 scoreboard sync 与 rebuild fallback
  - `submission_scoring.go` 仅保留入口编排
- Focused 验证完成：
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `timeout 180s env GOMAXPROCS=2 go -C /home/azhi/workspace/projects/ctf/code/backend test -p 1 -parallel 1 ./internal/app -run 'TestCompositionModulesExposeContracts' -count=1`
