# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 7

## Existing code searched
- `code/backend/internal/app/full_router_awd_state_matrix_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouterteacherauthoring/authoring_flow.go`
- `code/backend/tests/system/http/fullrouteradmin/admin_flow.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/fullrouterteacherauthoring/authoring_flow.go` 已经证明可以把长 HTTP 场景 owner 迁出，同时把 DB / 文件系统 seed 留在 `internal/app`。
- `full_router_awd_state_matrix_test.go` 里的 AWD 场景既有纯 HTTP 状态矩阵，也有少量 contest / instance seed，适合继续沿用 driver + callback 的模式。

## Decision
refactor_existing

## Reason
`full_router_awd_state_matrix` 的最小安全切片仍然是先迁断言 owner：

- 在 `code/backend/tests/system/http/fullrouterawdstate/` 建立可导入的场景断言 package
- `internal/app/full_router_awd_state_matrix_test.go` 保留 AWD seed、contest fixture 和少量 DB 结果校验
- 先不抽 `newFullRouterTestEnv`、AWD contest/service fixture owner 或 proxy 相关共享 helper

这样能继续缩小 `internal/app` 的系统测试 owner，同时避免本轮 scope 扩成 AWD 测试基建重做。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase7.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase7-plan.md`
- `code/backend/internal/app/full_router_awd_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouterawdstate/*.go`
- `code/backend/tests/README.md`

## After implementation
- `full_router_awd_state_matrix` 的核心 HTTP 场景断言 owner 不再放在 `internal/app`。
- `internal/app/full_router_awd_state_matrix_test.go` 收成 AWD seed、contest fixture 和兼容入口。
- full router fixture owner 仍暂留 `internal/app/full_router_integration_test.go`，后续继续按切片迁移。
