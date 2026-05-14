# Reuse Decision

## Change type

service / repository / port / contracts / readmodel

## Existing code searched

- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/challenge/contracts/contracts.go`
- `code/backend/internal/module/challenge/contracts/context_contract_test.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/challenge/infrastructure/repository_test.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/advice/advice_test.go`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`

## Similar implementations found

- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/teaching/advice/advice.go`

## Decision

refactor_existing

## Reason

这次不是新增一条推荐链路，也不是把推荐逻辑拆成新的模块，而是在现有 `assessment -> challenge` 推荐查询契约上补齐缺失的 difficulty owner，并让“健康学生进入进阶推荐”仍然走同一个 advice owner。

最小正确方案是：

- 继续复用 `teaching/advice` 作为推荐目标 owner，不在 `assessment` 里再造一套 progression 判断
- recommendation snapshot 补齐“维度 + 已解难度覆盖”事实，仍沿用现有教学事实快照，不新开独立推荐 read model
- 把 `challenge` 侧 `FindPublishedForRecommendation(...)` 从“只按维度筛题”改成“维度 + 目标难度带排序”
- 保留当前 `difficulty_band` 与实际题目 `difficulty` 并存的 DTO，不新造第二套返回结构
- 继续复用 `challenge` 仓储现有 category / knowledge tag 命中逻辑，不重复造推荐候选读取器

这样可以在不打散既有推荐链路的前提下，让难度带真正进入查询 owner，并保留“题库里最接近候选”这一降级语义。

## Files to modify

- `.harness/reuse-decisions/recommendation-difficulty-band-owner.md`
- `docs/plan/impl-plan/2026-05-14-recommendation-difficulty-band-owner-implementation-plan.md`
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/advice/advice_test.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/challenge/contracts/contracts.go`
- `code/backend/internal/module/challenge/contracts/context_contract_test.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/challenge/infrastructure/repository_test.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `docs/architecture/features/教学复盘建议生成架构.md`

## After implementation

- 推荐查询会显式消费目标 `difficulty_band`
- `difficulty_band` 继续表示目标训练带宽，`difficulty` 继续表示候选题实际难度
- 没有弱项或补样本目标时，推荐链路允许切到 progression 语义，而不是直接返回空
- 当题库里没有同维度、同难度题时，候选仍按最接近难度优先返回，而不是退回完全不受控的全局升序
