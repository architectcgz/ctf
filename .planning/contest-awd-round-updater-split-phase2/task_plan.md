# Task Plan

## Goal

继续收口 `contest` AWD round updater，把 `application/jobs/awd_rounds.go` 从单文件拆成更清晰的轮次规划与 flag 同步结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD round updater 职责 | completed | 已确认单文件同时承载轮次规划、落库、round lock、flag 同步与团队/题目装载 helper |
| 2. 拆分文件结构 | completed | 已拆为 `awd_rounds.go`、`awd_round_plan.go`、`awd_round_flag_sync.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 定向测试 |

## Acceptance Checks

- `awd_rounds.go` 不再承载混杂 AWD round updater 逻辑
- 轮次规划/落库 与 flag 同步/support 拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外接口
- 仅改善 AWD round updater 内部文件边界，为后续继续收口 contest jobs 留下更清晰结构
