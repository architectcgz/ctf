# Task Plan

## Goal

继续收口 `contest` infrastructure，把 `infrastructure/team_membership_repository.go` 从单文件拆成 team membership 事务流程与 registration 绑定 support 两段，保持 `TeamRepository` 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team membership repository 职责 | completed | 已确认单文件同时承载 team membership 事务流程与 registration/team 绑定 support |
| 2. 拆分文件结构 | completed | 已保留 `team_membership_repository.go` 承载 team membership 事务流程，并新增 `team_registration_binding.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `team_membership_repository.go` 不再混载 team membership 事务流程与 registration/team 绑定 support
- registration/team 绑定 helper 拆到独立 support 文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `TeamRepository` 对外接口与成员管理行为
- 仅改善 `contest` team membership infrastructure 文件边界，为后续继续收口 team repository 链路留下更清晰结构
