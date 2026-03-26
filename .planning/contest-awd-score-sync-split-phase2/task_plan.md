# Task Plan

## Goal

继续收口 `contest` AWD infrastructure，把 `infrastructure/awd_score_sync.go` 从单文件拆成更清晰的分数重算、缓存重建与解析 helper 结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD score sync 职责 | completed | 已确认单文件同时承载 score recalc、scoreboard cache rebuild 与 source/check parse helper |
| 2. 拆分文件结构 | completed | 已拆为 `awd_score_sync.go`、`awd_score_recalc.go`、`awd_scoreboard_cache.go`、`awd_score_sync_support.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 定向测试 |

## Acceptance Checks

- `awd_score_sync.go` 不再承载混杂 AWD score sync 逻辑
- score recalc / scoreboard cache / parse helper 拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 AWD repository 对外接口
- 仅改善 AWD infrastructure 内部文件边界，为后续继续收口 contest repository 层留下更清晰结构
