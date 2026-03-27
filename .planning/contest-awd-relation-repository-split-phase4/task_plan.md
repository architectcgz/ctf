# Task Plan

## Goal

继续收口 `contest` AWD infrastructure，把 `infrastructure/awd_relation_repository.go` 从单文件拆成 contest/challenge relation 与 team/member relation 两段，保持 `AWDRepository` 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd relation repository 职责 | completed | 已确认单文件同时承载 contest/challenge relation 与 team/member/registration relation |
| 2. 拆分文件结构 | completed | 已拆为 `awd_contest_relation_repository.go` 与 `awd_team_relation_repository.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_relation_repository.go` 不再混载 contest/challenge relation 与 team/member/registration relation
- 两类 AWD relation repository 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDRepository` 对外接口与 relation 查询行为
- 仅改善 `contest` AWD infrastructure 文件边界，为后续继续收口 AWD repository 链路留下更清晰结构
