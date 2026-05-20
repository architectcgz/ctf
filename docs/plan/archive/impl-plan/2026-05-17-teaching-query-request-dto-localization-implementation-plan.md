# teaching query request DTO 模块内化实现方案

## Objective

把教师侧 class / student / evidence / attack session 这批 request DTO 从全局 `internal/dto/teacher.go` 收回 `teaching_query` owning module，明确 `api/http` 与 `application/queries` 的 owner 边界，同时保持现有教师端接口的 query 参数、校验口径和 response shape 不变。

## Non-goals

- 不改教师侧 response DTO owner；这刀只处理 request/input
- 不调整 `GET /api/v1/teacher/*` 的路径、参数名、分页默认值和权限校验
- 不扩到 `teacher.go` 里其余 overview / class / student response DTO 的去全局化

## Inputs

- `docs/architecture/backend/04-api-design.md`
- `docs/plan/impl-plan/2026-05-17-teacher-aggregate-dto-localization-implementation-plan.md`
- `code/backend/internal/module/teaching_query/api/http/handler.go`
- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/teaching_query/application/queries/service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/dto/teacher.go`

## Task Slices

1. 新增 teaching query 本地 request / input 类型
   - `teaching_query/api/http`：承接 query 绑定 DTO
   - `teaching_query/application/queries`：承接纯 input 类型

2. 收口 handler -> application 边界
   - handler 绑定模块内 request DTO
   - 通过薄 mapper 映射到 query input
   - service / class insight / student review 接口全部改吃模块内 input

3. 清理全局 DTO 与补充验证
   - 从 `internal/dto/teacher.go` 删除已无 owner 的 6 个 request 类型
   - 更新 teaching query 相关单测、handler 绑定测试和 API 设计文档

## Expected Changes

- `code/backend/internal/module/teaching_query/api/http/handler.go`
- `code/backend/internal/module/teaching_query/api/http/request_types.go`
- `code/backend/internal/module/teaching_query/api/http/request_mapper.go`
- `code/backend/internal/module/teaching_query/api/http/handler_test.go`
- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/teaching_query/application/queries/input_types.go`
- `code/backend/internal/module/teaching_query/application/queries/service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service_test.go`
- `code/backend/internal/dto/teacher.go`
- `docs/architecture/backend/04-api-design.md`
- `docs/reviews/backend/2026-05-17-teaching-query-request-dto-localization-review.md`

## Validation

- `go test ./internal/module/teaching_query/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_TeacherAccessAndRecommendationStateMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestFullRouter_InstanceHintAndProxyStateMatrix' -count=1`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- `teaching_query/api/http` 是否成为这 6 个 request DTO 的唯一 HTTP owner
- application input 是否已经脱离 `form` / `binding` / `time_format` 语义
- handler 是否只做 trust boundary 绑定，不再把 HTTP DTO 直接下传到 query service
- 删除全局 request 类型后，教师侧 class insight / student review 行为与现有 API 契约是否保持一致
