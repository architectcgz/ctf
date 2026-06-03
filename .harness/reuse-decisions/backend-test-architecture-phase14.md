# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 14

## Existing code searched
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/full_router_*_test.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `newFullRouterTestEnv`、`performFullRouterRequest` 以及一组 seed helper 已被多个 `full_router_*_test.go` 文件共享调用。
- 当前这些 helper 仍集中堆在 `full_router_integration_test.go`，导致该文件虽然只剩一个测试，但体量仍然很大。

## Decision
refactor_existing

## Reason
phase 14 改为 shared helper owner 拆分：

- 新增独立的 `internal/app/full_router_test_helpers_test.go`
- 把通用 env / config / seed / request helper 从 `full_router_integration_test.go` 挪过去
- 保留 `full_router_integration_test.go` 里的 router/module wiring 测试和 access-specific helper

这样能显著压缩 `full_router_integration_test.go`，同时不改变 helper 的包内可见性和调用方式。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase14.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase14-plan.md`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_test_helpers_test.go`

## After implementation
- `full_router_integration_test.go` 不再承担 full router 通用 fixture / seed owner。
- 共享 helper 集中到专门的 test helper 文件，供 admin/teacher/contest/access/state 测试继续复用。
