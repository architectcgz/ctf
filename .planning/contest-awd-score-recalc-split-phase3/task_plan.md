# Task Plan

## Goal

继续收口 `contest` AWD infrastructure，把 `infrastructure/awd_score_recalc.go` 从单文件拆成重算主流程与积分归并 support 两段，保持 AWD 官方积分口径不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd score recalc 职责 | completed | 已确认单文件同时承载重算/同步入口与 defense/attack 积分归并规则 |
| 2. 拆分文件结构 | completed | 已保留 `awd_score_recalc.go` 为主流程，并新增 `awd_score_recalc_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_score_recalc.go` 不再混载重算入口与积分归并 support
- defense/attack 积分归并与 official totals 判定拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 AWD 官方积分重算行为
- 仅改善 `contest` AWD 记分 infrastructure 文件边界，为后续继续收口 AWD score 链路留下更清晰结构
