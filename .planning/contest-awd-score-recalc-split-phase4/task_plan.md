# Task Plan

## Goal

继续收口 `contest` AWD score recalc，把 `infrastructure/awd_score_recalc.go` 从单文件拆成入口编排、数据加载与分数写回三段，保持 `AWDRepository` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd score recalc 职责 | completed | 已确认单文件同时承载 contest/team/service/attack 读取与逐队写回更新 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_score_recalc.go`、`awd_score_recalc_loader.go`、`awd_score_recalc_writeback.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_score_recalc.go` 不再混载读取与写回细节
- contest/team/service/attack 读取下沉到 loader 文件
- team 总分与 `last_solve_at` 写回下沉到 writeback 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRepository` 对外接口与分数重算行为
- 仅改善 AWD score recalc 内部文件边界，为后续继续收口 contest infrastructure 留下更清晰结构
