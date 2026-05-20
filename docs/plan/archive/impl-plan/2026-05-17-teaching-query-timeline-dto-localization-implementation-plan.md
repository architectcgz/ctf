# teaching query timeline DTO 模块内化实现方案

## Objective

把教师侧学生时间线响应 DTO 从 `internal/dto/timeline.go` 收回 `teaching_query/application/queries`，保持 `GET /api/v1/teacher/students/:id/timeline` 外部契约不变。

## Non-goals

- 不处理 `teacher.go` 里的 class / overview / recommendation / evidence / attack session 类型
- 不处理 `teacher_awd_review.go`
- 不改教师侧路由、权限、JSON 字段、分页参数

## Inputs

- `docs/architecture/backend/04-api-design.md`
- `docs/plan/impl-plan/2026-05-17-teacher-aggregate-dto-localization-implementation-plan.md`
- `code/backend/internal/module/teaching_query/application/queries/{contracts.go,student_review_service.go,response_mapper.go}`
- `code/backend/internal/dto/timeline.go`

## Task Slices

1. 在 `teaching_query/application/queries` 定义本地 timeline response 类型
   - 保持 `events`、`timestamp`、`is_correct`、`points`、`detail` 字段不变

2. 收口 student review query 接口与 mapper
   - `StudentReviewService.GetStudentTimeline` 改为只返回模块内 `TimelineResp`
   - 重新生成 `response_mapper_gen.go`

3. 调整 app 集成测试解码类型
   - `full_router_state_matrix_integration_test.go` 改为使用 `teaching_query/application/queries.TimelineResp`

4. 删除废弃全局 DTO
   - 若 `internal/dto/timeline.go` 无剩余引用则删除

## Expected Changes

- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/application/queries/timeline_types.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/timeline.go`

## Validation

- `go test ./internal/module/teaching_query/... -count=1`
- `go test ./internal/app -run 'TestFullRouter_TeacherAccessAndRecommendationStateMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`

## Review Focus

- timeline DTO owner 是否已经回到 `teaching_query/application/queries`
- 是否只改了 owner，没有改教师侧时间线外部 HTTP 契约
- app 集成测试与 mapper 是否已经不再引用全局 `dto.TimelineResp`
