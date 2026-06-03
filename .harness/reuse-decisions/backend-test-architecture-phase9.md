# Reuse Decision

## Change type
backend test refactor / test architecture migration phase 9

## Existing code searched
- `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/tests/system/http/fullrouterconteststate/contest_state.go`
- `code/backend/tests/system/http/fullrouterteacherauthoring/authoring_flow.go`
- configured search roots:
  - `code/backend/internal/module`
  - `code/backend/internal/app/composition`
  - `code/backend/internal/model`
  - `code/backend/internal/app`
  - `code/backend/tests`

## Similar implementations found
- `tests/system/http/fullrouterconteststate/contest_state.go` 已经证明可以把多场景 HTTP 状态矩阵 owner 迁出，同时把 seed 留在 `internal/app`。
- `full_router_teacher_state_matrix_test.go` 里只有 teacher access / recommendation 与 challenge writeup semantics 两组场景适合本轮迁移；`TeacherAWDReviewExportStateMatrix` 仍有既有 PDF 断言失败，不适合和架构迁移混做一刀。

## Decision
refactor_existing

## Reason
`full_router_teacher_state_matrix` 本轮选择更小且更安全的切片：

- 在 `code/backend/tests/system/http/fullrouterteacherstate/` 建立可导入的场景断言 package
- 只迁 `TeacherAccessAndRecommendationStateMatrix` 与 `ChallengeWriteupsUseCommunitySemantics`
- `TeacherAWDReviewExportStateMatrix` 暂留 `internal/app`，避免把已知既有失败和 owner 迁移耦合到同一轮

这样能继续缩小 `internal/app` 的系统测试 owner，同时不把本轮 scope 扩成“既拆测试又修 PDF 导出”的混合任务。

## Files to modify
- `.harness/reuse-decisions/backend-test-architecture-phase9.md`
- `docs/plan/impl-plan/2026-06-03-backend-test-architecture-phase9-plan.md`
- `code/backend/internal/app/full_router_teacher_state_matrix_test.go`
- `code/backend/tests/system/http/fullrouterteacherstate/*.go`
- `code/backend/tests/README.md`

## After implementation
- `full_router_teacher_state_matrix` 中非 AWD review 的核心 HTTP 场景断言 owner 不再放在 `internal/app`。
- `TeacherAWDReviewExportStateMatrix` 继续留在 `internal/app`，后续单独处理其既有失败和迁移。
- full router fixture owner 仍暂留 `internal/app/full_router_integration_test.go`，后续继续按切片迁移。
