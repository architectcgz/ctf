# Reuse Decision

## Change type
entity / repository / testsupport / app-test / model localization

## Existing code searched
- `code/backend/internal/model/contest_status_transition.go`
- `code/backend/internal/module/contest/...`
- `code/backend/internal/app/full_router_integration_test.go`
- `.harness/reuse-decisions/*contest*status*transition*`

## Similar implementations found
- `code/backend/internal/module/contest/entity/announcement.go`
- `code/backend/internal/module/ops/entity/audit_log.go`
- `code/backend/internal/module/practice/entity/user_score.go`

## Decision
refactor_existing

## Reason
`ContestStatusTransition` 是 `contest` 模块状态机持久化记录，只服务于 contest 的状态推进、手动冻结/解冻和副作用重放，不适合继续留在全局 `internal/model`。最小正确方案是把 GORM 实体收回 `internal/module/contest/entity`，保持表名、字段和状态机行为不变。

非目标：本刀不处理 `Contest`、`ContestRegistration`、`ContestChallenge`。

## Files to modify
- `code/backend/internal/model/contest_status_transition.go`
- `code/backend/internal/module/contest/entity/status_transition.go`
- `code/backend/internal/module/contest/infrastructure/contest_status_transition_repository.go`
- `code/backend/internal/module/contest/infrastructure/contest_status_transition_repository_test.go`
- `code/backend/internal/module/contest/infrastructure/contest_status_update_repository_test.go`
- `code/backend/internal/module/contest/application/jobs/status_updater_test.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_service_test.go`
- `code/backend/internal/module/contest/application/commands/contest_service_test.go`
- `code/backend/internal/module/contest/testsupport/db.go`
- `code/backend/internal/app/full_router_integration_test.go`

## After implementation
- 删除 `internal/model/contest_status_transition.go`
- 后续再单独处理其他 contest 共享实体
