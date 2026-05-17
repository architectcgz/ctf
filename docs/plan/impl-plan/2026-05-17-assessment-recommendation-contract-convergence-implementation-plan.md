# assessment recommendation contract 收口实现方案

## Objective

把 `assessment/contracts.RecommendationProvider` 从直接返回 `internal/dto` 收口为模块内 contract 类型，并把 HTTP DTO 映射压回 `assessment/api/http` 与 `teaching_query` 的消费边界。

## Non-goals

- 不迁移 `assessment` 里 skill profile、report、teacher AWD review 的 DTO 结构
- 不调整推荐接口路径、查询参数或外部 JSON 字段
- 不重做 `teaching_query` 的整体 contract 体系

## Inputs

- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/teaching_query/application/queries/{class_insight_service.go,student_review_service.go,response_mapper.go}`

## Task Slices

1. 在 `assessment/contracts` 定义 recommendation contract 类型
   - 新增 `Recommendation`、`ChallengeRecommendation`、`RecommendationWeakDimension`
   - `RecommendationProvider` 改为返回 contract 类型
   - 验证：`contracts` 不再 import `internal/dto`

2. 收口 `assessment` application 输出
   - `recommendation_service.go` 改为返回 contract 类型
   - 更新 recommendation 相关测试
   - 验证：`go test ./internal/module/assessment/...`

3. 把 DTO 映射压回消费边界
   - `assessment/api/http` 把 contract 映射回 `dto.RecommendationResp`
   - `teaching_query` 把 assessment contract 映射回教师端 recommendation DTO
   - 验证：`go test ./internal/module/assessment/... ./internal/module/teaching_query/...`

4. 守住新边界
   - 更新 `assessment/architecture_test.go`，禁止 `contracts` 回流 `internal/dto`
   - 验证：`go test ./internal/module/assessment/...`

## Expected Changes

- `code/backend/internal/module/assessment/contracts/`
- `code/backend/internal/module/assessment/application/queries/`
- `code/backend/internal/module/assessment/api/http/`
- `code/backend/internal/module/assessment/architecture_test.go`
- `code/backend/internal/module/teaching_query/application/queries/`
- `docs/plan/impl-plan/2026-05-17-assessment-recommendation-contract-convergence-implementation-plan.md`

## Compatibility

- 外部 `GET /api/v1/recommendations` 响应结构保持不变
- `teaching_query` 对教师端 recommendation 的输出结构保持不变
- 这次只改变 `assessment` 对外 contract 和内部映射 owner，不改变推荐算法与缓存语义

## Validation

- `go test ./internal/module/assessment/...`
- `go test ./internal/module/teaching_query/...`
- `go test ./internal/app/... -run 'TestCompositionModulesExposeContracts|TestTeachingQueryModuleUsesTypedDeps'`

## Review Focus

- `assessment/contracts` 是否已经不再依赖 `internal/dto`
- `assessment/api/http` 是否成为 recommendation HTTP DTO 的唯一 owner
- `teaching_query` 是否只消费 `assessment` contract，而不是继续依赖 `dto.RecommendationResp`

## Rollback

- 如果 contract 替换引起 teaching query 编译或映射回归，可先回退 `RecommendationProvider` 的签名与对应 mapper，再按更小切片重做
