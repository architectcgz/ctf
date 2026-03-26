# Task Plan

## Goal

继续收口 `contest` AWD 写侧 helper，把 `application/commands/awd_support.go` 从单文件拆成更清晰的 support 分组，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD 写侧 helper 职责 | completed | 已确认 `awd_support.go` 同时承载校验、当前轮次解析与 flag/grace period helper |
| 2. 拆分文件结构 | completed | 已拆为 `awd_support.go`、`awd_validation_support.go`、`awd_round_support.go`、`awd_flag_support.go` |
| 3. focused 验证 | completed | 已运行 `./internal/module/contest/...` 定向测试 |

## Acceptance Checks

- `awd_support.go` 不再承载混杂 helper 实现
- AWD 写侧校验、轮次解析、flag helper 拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 AWD commands 对外接口
- 仅改善 commands 层 helper 边界，为后续继续拆 AWD 写侧主流程留出更清晰落点
