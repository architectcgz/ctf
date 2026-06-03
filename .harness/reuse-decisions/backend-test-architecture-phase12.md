# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 12

## Existing code searched
- `code/backend/internal/app/full_router_admin_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/fullrouteradmin/admin_flow.go` 已经承接多个 `full_router_admin` 场景断言 owner，适合继续扩展到 admin ops / notification。
- `internal/app/full_router_state_matrix_integration_test.go` 里已有 `performFullRouterMultipartRequest` 和 `receiveFullRouterWSMessageByType`，本轮通过 driver / package-local helper 复用其行为，不改 helper owner。

## Decision
refactor_existing

## Reason
phase 12 继续做单一 owner 迁移：

- 在既有 `code/backend/tests/system/http/fullrouteradmin/` 下新增 admin ops / notification 场景断言
- `internal/app/full_router_admin_state_matrix_test.go` 收敛到 page size 小回归和最小 glue
- 不在本轮抽 multipart / websocket helper 到共享 testutil

这样可以把 admin 状态矩阵的大场景基本迁空，同时避免 helper owner 抽取扩大范围。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase12.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase12-plan.md`
- `code/backend/internal/app/full_router_admin_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- `code/backend/tests/README.md`

## After implementation
- `AdminOpsAndNotificationStateMatrix` 的核心 HTTP 场景断言 owner 不再放在 `internal/app`。
- `full_router_admin_state_matrix_test.go` 只剩 page size 小回归和 glue。
- multipart / websocket helper 是否抽共享 testutil，后续再单独切片处理。
