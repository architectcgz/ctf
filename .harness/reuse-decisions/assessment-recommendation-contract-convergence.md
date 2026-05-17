# Reuse Decision

## Change type

service / cache / contract / handler / mapper

## Existing code searched

- `code/backend/internal/module/assessment/contracts/*.go`
- `code/backend/internal/module/assessment/application/queries/*.go`
- `code/backend/internal/module/assessment/api/http/*.go`
- `code/backend/internal/module/assessment/ports/*.go`
- `code/backend/internal/module/assessment/infrastructure/state_store.go`
- `code/backend/internal/module/teaching_query/application/queries/*.go`
- `code/backend/internal/dto/assessment.go`
- `code/backend/internal/dto/teacher.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`

## Similar implementations found

- `assessment/api/http/request_mapper.go` 已经是 assessment HTTP 边界的 DTO 映射落点，适合继续承接 recommendation response 映射
- `teaching_query/application/queries/response_mapper.go` 已经承担教师端聚合查询结果到 DTO 的映射，适合直接复用为 assessment recommendation contract 的消费边界
- `assessment/contracts` 已经是 assessment 对外稳定能力的 owner，不需要再新增全局 shared recommendation 包

## Decision

refactor_existing

## Reason

这次不是新增新的推荐模块、并行 DTO 包或新的跨模块 shared transport，而是在现有 `assessment` 模块内把 recommendation 的稳定 contract 与 HTTP DTO 拆开：

- `assessment/contracts` 新增 recommendation contract 结构，作为跨模块稳定输出
- `assessment/application/queries` 和 recommendation cache 改为围绕 contract 类型工作，不再直接暴露 `internal/dto`
- `assessment/api/http` 负责把 contract 映射回学生端 recommendation HTTP DTO
- `teaching_query/application/queries` 负责把 assessment contract 映射回教师端 recommendation DTO

这样复用了现有模块边界、现有 handler、现有 mapper 生成方式和现有 teaching query 消费面，没有引入新的并行抽象，也避免继续扩大 `internal/dto` 这个全局共享桶。

## Files to modify

- `.harness/reuse-decisions/assessment-recommendation-contract-convergence.md`
- `docs/plan/impl-plan/2026-05-17-assessment-recommendation-contract-convergence-implementation-plan.md`
- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/module/assessment/api/http/response_mapper_gen.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/architecture_test.go`
- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/infrastructure/state_store.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/assessment/ports/state_store_context_contract_test.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service_test.go`

## After implementation

- `assessment/contracts.RecommendationProvider` 不再直接返回 `internal/dto`
- recommendation cache 的持久化形状跟随 `assessment/contracts`，而不是继续绑定 HTTP DTO
- 学生端和教师端 recommendation 输出结构保持不变，但映射 owner 回到各自的 API / 消费边界
