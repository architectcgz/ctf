# Task Plan

## Goal

继续收口 `contest` 内部拆分，把 AWD 读侧查询服务从单文件拆成更清晰的 `service / query / support` 结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD 读侧职责 | completed | 已确认 `application/queries/awd_service.go` 同时承载构造、查询流程与校验/装载辅助 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_service.go`、`awd_query.go`、`awd_support.go` |
| 3. focused 验证 | completed | 已运行 `./internal/module/contest/...` 定向测试 |

## Acceptance Checks

- `contest/application/queries/awd_service.go` 只保留 service 定义与构造器
- AWD 读侧查询流程与 support helper 拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 AWD 对外 API
- 仅改善 `contest` AWD 读侧文件边界，为后续继续拆 AWD 逻辑留出更清晰落点
