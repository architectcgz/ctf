# Progress

## 2026-03-26

- 启动 `contest-awd-score-sync-split-phase2`，目标是继续拆 `contest` AWD score sync 主流程文件。
- 盘点确认 `infrastructure/awd_score_sync.go` 同时承载三类职责：
  - 官方分数重算与累计逻辑
  - 排行榜缓存重建逻辑
  - source/check 解析与时间解析 helper
- 已完成文件拆分：
  - `awd_score_sync.go` 收缩为 package 占位
  - `awd_score_recalc.go` 承载分数重算与累计 helper
  - `awd_scoreboard_cache.go` 承载排行榜缓存重建逻辑
  - `awd_score_sync_support.go` 承载 parse / normalize helper
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
