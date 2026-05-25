# Reuse Decision

## Change type

service / repository / port / contracts / command

## Existing code searched

- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/classwindow/window.go`
- `code/backend/internal/module/challenge/contracts/*.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/teaching_query/ports/query.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/architecture_test.go`
- `docs/plan/impl-plan/2026-05-14-recommendation-difficulty-band-owner-implementation-plan.md`
- `.harness/reuse-decisions/recommendation-difficulty-band-owner.md`
- `.harness/reuse-decisions/teaching-query-identity-user-lookup.md`

## Similar implementations found

- `code/backend/internal/module/challenge/contracts/skill_dimension.go`
- `code/backend/internal/module/challenge/contracts/persistence.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/teaching/classwindow/window.go`

## Decision

refactor_existing

## Reason

这次不是新增 teaching 模块，也不是重写推荐与教师查询链路，而是在现有 shared-kernel + module 分层上把三处边界污染收回来。

最小正确方案是：

- 复用现有维度/难度语义，抽成真正共享的 taxonomy source，而不是在 `teaching` 或 `assessment` 再各造一套枚举
- 让所有共享语义消费者直接依赖 `internal/shared/taxonomy`，而不是继续经由 `challenge/contracts` 中转
- 保留 `teaching_query/ports` 里的读模型名字与接口形状，只把 GORM tag 下沉到 `infrastructure` row
- 保留 `classwindow.Parse(now, from, to)` 这个 API 形状，只移除零值 `now` 的隐式兜底

这样可以先把依赖方向、共享语义 owner、端口纯度和规则确定性收紧，而不把本轮任务膨胀成目录大迁移。

## Files to modify

- `.harness/reuse-decisions/teaching-shared-kernel-boundary.md`
- `docs/plan/archive/impl-plan/2026-05-19-teaching-shared-kernel-boundary-implementation-plan.md`
- `code/backend/internal/shared/taxonomy/*.go`
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/advice/advice_test.go`
- `code/backend/internal/teaching/classwindow/window.go`
- `code/backend/internal/teaching/classwindow/window_test.go`
- `code/backend/internal/module/challenge/contracts/persistence.go`
- `code/backend/internal/module/challenge/architecture_test.go`
- `code/backend/internal/module/teaching_query/ports/query.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/architecture_test.go`
- `code/backend/internal/module/assessment/**/*.go`
- `code/backend/internal/module/practice/**/*.go`
- `code/backend/internal/module/runtime/**/*.go`
- `code/backend/internal/module/contest/**/*.go`
- `code/backend/internal/app/**/*.go`
- `code/backend/cmd/seed-teaching-review-data/*.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`

## After implementation

- `internal/teaching` 不再直接依赖 `challenge/contracts`
- `challenge/contracts` 不再重导出共享 taxonomy
- `teaching_query/ports` 不再声明任何 GORM tag
- 时间窗默认行为仍保留，但必须由调用方显式提供 `now`
