# Task Plan

## Goal

继续收口 `contest` AWD command support，把 `application/commands/awd_validation_support.go` 从单文件拆成 contest/round 校验、resource 校验加载与 team 归属解析三段，保持 `AWDService` 校验行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd validation support 职责 | completed | 已确认单文件同时承载 contest/round 校验、resource 校验与 user team 解析 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_resource_validation_support.go` 与 `awd_team_validation_support.go`，`awd_validation_support.go` 保留 contest/round 校验 |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_validation_support.go` 不再混载 resource 校验与 user team 解析
- contest/challenge/team 资源校验和加载拆到独立 support 文件
- user team 归属解析拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDService` 校验 helper 的对外函数签名与行为
- 仅改善 AWD validation support 文件边界，为后续继续收口 AWD command support 留下更清晰结构
