# Task Plan

## Goal

继续收口 `contest` team membership infrastructure，把 `infrastructure/team_membership_repository.go` 从单文件拆成 team 生命周期事务与成员加入/离队事务两段，保持 `TeamRepository` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 team membership repository 职责 | completed | 已确认单文件同时承载 Create/Delete 生命周期事务与 Add/Remove 成员事务 |
| 2. 拆分文件结构 | completed | 已拆为 `team_membership_lifecycle_repository.go` 与 `team_membership_repository.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `team_membership_repository.go` 不再混载生命周期事务与成员事务
- CreateWithMember/DeleteWithMembers 拆到独立 lifecycle 文件
- AddMemberWithLock/RemoveMember 保留成员事务文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `TeamRepository` 对外接口与事务行为
- 仅改善 team membership repository 文件边界，为后续继续收口 team infrastructure 留下更清晰结构
