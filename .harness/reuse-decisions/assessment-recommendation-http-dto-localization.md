# Reuse Decision

## Change type

handler / mapper / api

## Existing code searched

- `code/backend/internal/module/assessment/api/http/*.go`
- `code/backend/internal/module/identity/api/http/*.go`
- `code/backend/internal/module/auth/api/http/*.go`
- `code/backend/internal/module/teaching_query/api/http/*.go`
- `code/backend/internal/dto/recommendation.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/04-api-design.md`

## Similar implementations found

- `assessment/api/http/request_mapper.go` 已经承担 assessment 模块的 HTTP DTO 到 application input 映射
- `assessment/api/http/response_mapper.go` 已经是 recommendation contract 到 HTTP response 的既有落点，只差把 response 结构体 owner 从全局 `internal/dto` 收回模块内
- `teaching_query/api/http` 仍保留教师端 DTO 输出，本次不需要顺手迁移教师 recommendation DTO

## Decision

refactor_existing

## Reason

这次不是新增新的 recommendation handler 或第二套响应契约，而是在既有 `assessment/api/http` 边界上把学生端 recommendation 的 HTTP response 类型本地化：

- 沿用现有 `assessment/api/http` handler 和 goverter response mapper
- 把 `RecommendationResp`、`ChallengeRecommendation`、`RecommendationWeakDimension` 从全局 `internal/dto` 收回 `assessment/api/http`
- 保持外部 JSON 字段和路由不变，只调整 response type owner

这样能延续前一刀已经完成的 contract 收口，让 `assessment` recommendation 的跨模块 contract 和 HTTP DTO owner 都回到模块边界内，而不是继续依赖全局 DTO 桶。

## Files to modify

- `.harness/reuse-decisions/assessment-recommendation-http-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-17-assessment-recommendation-http-dto-localization-implementation-plan.md`
- `docs/architecture/backend/04-api-design.md`
- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/module/assessment/api/http/response_mapper_assign.go`
- `code/backend/internal/module/assessment/api/http/response_mapper_gen.go`
- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## After implementation

- `GET /api/v1/users/me/recommendations` 的 HTTP response DTO owner 回到 `assessment/api/http`
- `assessment` recommendation 链路不再需要全局 `internal/dto/recommendation.go` 参与模块内 HTTP 映射
- 若后续确认没有剩余引用，再单独删除全局 `recommendation.go`
