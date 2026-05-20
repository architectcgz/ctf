# assessment teacher AWD review request DTO 模块内化实现方案

## Objective

把教师侧 AWD review 的 HTTP 请求 DTO 从 `internal/dto/teacher_awd_review.go` 收回 `assessment/api/http`，保持 `GET /api/v1/teacher/awd/reviews` 与 `GET /api/v1/teacher/awd/reviews/:id` 外部契约不变。

## Non-goals

- 不处理 `teacher_awd_review.go` 里的响应 DTO
- 不改 AWD review export / report 的响应或导出链
- 不处理 `teacher.go`、`timeline.go` 或其他教师侧聚合 DTO

## Inputs

- `docs/architecture/backend/04-api-design.md`
- `docs/plan/impl-plan/2026-05-17-teacher-aggregate-dto-localization-implementation-plan.md`
- `code/backend/internal/module/assessment/api/http/{teacher_awd_review_handler.go,request_mapper.go,request_mapper_gen.go}`
- `code/backend/internal/dto/teacher_awd_review.go`

## Task Slices

1. 在 `assessment/api/http` 定义本地 AWD review request DTO
   - `TeacherAWDReviewContestQuery`
   - `GetTeacherAWDReviewArchiveReq`

2. 收口 handler 与 request mapper
   - `TeacherAWDReviewHandler` 改为绑定本地请求类型
   - 重新生成 `request_mapper_gen.go`

3. 调整路由测试
   - `router_test.go` 改为反射本地 request DTO，锁住 `round`、`team_id`、`page_size` 等 query tag

4. 从全局 DTO 文件删除已迁移请求类型
   - 仅删除 request DTO，保留响应 DTO 给后续切片处理

## Expected Changes

- `code/backend/internal/module/assessment/api/http/teacher_awd_review_request_types.go`
- `code/backend/internal/module/assessment/api/http/teacher_awd_review_handler.go`
- `code/backend/internal/module/assessment/api/http/request_mapper.go`
- `code/backend/internal/module/assessment/api/http/request_mapper_gen.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/dto/teacher_awd_review.go`

## Validation

- `go generate ./internal/module/assessment/api/http`
- `go test ./internal/module/assessment/api/http -count=1`
- `go test ./internal/app -run 'TestTeacherAWDReviewArchiveReqUsesPlannedQueryParams|TestTeacherAWDReviewContestQueryUsesPlannedQueryParams|TestTeacherAWDReviewServiceInvalidRoundUsesRoundMessage' -count=1`

## Review Focus

- request DTO owner 是否已经回到 `assessment/api/http`
- 查询参数名和可选字段语义是否保持不变
- 全局 `teacher_awd_review.go` 是否只删了 request DTO，没有误碰响应链
