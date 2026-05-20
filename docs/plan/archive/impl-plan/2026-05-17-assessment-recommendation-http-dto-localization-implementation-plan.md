# assessment recommendation HTTP DTO 模块内化实现方案

## Objective

把学生端 recommendation 的 HTTP response DTO 从 `internal/dto` 收回 `assessment/api/http`，让 recommendation 这条链同时完成 contract owner 和 HTTP DTO owner 的模块内化。

## Non-goals

- 不迁移 `assessment` 的 report、skill profile、teacher AWD review DTO
- 不调整 `teaching_query` 教师端 recommendation DTO 归属
- 不新增、删除或修改 recommendation 的外部 JSON 字段、路由和查询参数
- 本刀不直接删除 `internal/dto/recommendation.go`

## Inputs

- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/04-api-design.md`
- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/dto/recommendation.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Task Slices

1. 在 `assessment/api/http` 定义 recommendation response 类型
   - 新增模块内 `RecommendationResp`、`ChallengeRecommendation`、`RecommendationWeakDimension`
   - 验证：字段名与现有 JSON tag 保持一致

2. 收口 mapper 与 handler
   - `response_mapper.go` 与生成文件改为输出模块内 response 类型
   - `handler.go` 继续返回现有 JSON 结构，但不再依赖全局 recommendation DTO
   - 验证：`go test ./internal/module/assessment/...`

3. 更新调用侧验证
   - 集成测试改为用模块内 response 类型解码学生端 recommendation
   - 文档补充“无外部契约变化，仅 owner 收口”的事实说明
   - 验证：`go test ./internal/app/... -run TestFullRouter_TeacherAccessAndRecommendationStateMatrix`

## Expected Changes

- `code/backend/internal/module/assessment/api/http/`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `docs/architecture/backend/04-api-design.md`
- `.harness/reuse-decisions/assessment-recommendation-http-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-assessment-recommendation-http-dto-localization-implementation-plan.md`

## Compatibility

- `GET /api/v1/users/me/recommendations` 的外部 JSON 字段保持不变
- `GET /api/v1/teacher/students/:id/recommendations` 不在本刀变更范围内

## Validation

- `go test ./internal/module/assessment/...`
- `go test ./internal/app/... -run TestFullRouter_TeacherAccessAndRecommendationStateMatrix`

## Review Focus

- recommendation HTTP DTO 是否已经回到 `assessment/api/http`
- 是否只改 owner，不改外部 JSON 契约
- 集成测试是否覆盖到学生端 recommendation 的真实解码路径

## Rollback

- 如果模块内 response 类型导致集成测试或 mapper 生成异常，可先回退 `assessment/api/http/response_mapper.go` 与集成测试解码类型，再重新切更小范围
