# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 11

## Existing code searched
- `code/backend/internal/app/full_router_admin_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/fullrouteradmin/admin_flow.go` 已经承接 `full_router_admin` 的 AWD control 和 publish request lifecycle 场景断言 owner。
- `full_router_admin_state_matrix_test.go` 剩余三个测试里，`TestFullRouter_AdminChallengeManagementStateMatrix` 是最适合单独迁移的长 HTTP 场景；`AdminOpsAndNotificationStateMatrix` 与 `AdminImagesCapsOversizedPageSize` 暂不与本轮混做。

## Decision
refactor_existing

## Reason
phase 11 继续沿用“先迁场景断言 owner，后拆 helper owner”的节奏：

- 在既有 `code/backend/tests/system/http/fullrouteradmin/` 下新增 challenge management 场景断言
- `internal/app/full_router_admin_state_matrix_test.go` 只保留 glue code、剩余 admin 场景和本地 helper
- 不在这一刀抽 `createPracticeSubmission`、实例 seed 或 DB 状态更新 helper

这样可以继续缩小 `internal/app` 的系统测试 owner，同时保持 review 面清晰。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase11.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase11-plan.md`
- `code/backend/internal/app/full_router_admin_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- `code/backend/tests/README.md`

## After implementation
- `AdminChallengeManagementStateMatrix` 的核心 HTTP 场景断言 owner 不再放在 `internal/app`。
- `full_router_admin_state_matrix_test.go` 暂时仍保留 admin ops/notification 场景和 page size 小测试。
- challenge 相关共享 helper 是否进一步抽到 testkit，后续单独切片处理。
