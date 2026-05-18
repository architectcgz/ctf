# teacher / submission / report DTO 模块内化实现方案

## Objective

把这批明显的历史全局 DTO 从 `internal/dto` 拆回 owning module：

- `teaching_query`：`teacher.go` 中的 class / overview / student review / evidence / attack session，以及 `timeline.go`
- `runtime`：`teacher.go` 中的 teacher instance query / item
- `assessment`：`teacher_awd_review.go` 与 `report.go`
- `practice`：`submission.go` 中的 submit flag / submission record
- `contest`：`submission.go` 中的 contest submit flag response

## Non-goals

- 不改教师侧、练习侧、竞赛侧已有路由、权限、JSON 字段、分页语义
- 不改 report export 的 HTTP 契约
- 不扩到 `challenge / contest / instance / awd` 其他全局 DTO

## Inputs

- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/teaching_query/**/*`
- `code/backend/internal/module/runtime/**/*`
- `code/backend/internal/module/assessment/**/*teacher_awd_review*`
- `code/backend/internal/module/practice/**/*submission*`
- `code/backend/internal/module/contest/**/*submission*`
- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/dto/timeline.go`
- `code/backend/internal/dto/teacher_awd_review.go`
- `code/backend/internal/dto/submission.go`
- `code/backend/internal/dto/report.go`

## Task Slices

1. `teaching_query`
   - 新增模块内 contract 与 `api/http` request / response DTO
   - `application/queries` 不再依赖全局 `dto.Teacher*` / `dto.TimelineResp`

2. `runtime`
   - 新增 teacher instance query / response 的模块内类型
   - `runtime/api/http` 与 application/query service 不再依赖 `dto.TeacherInstance*`

3. `assessment`
   - 新增 teacher AWD review 的模块内 contract 与 HTTP DTO
   - report archive / renderer 改用模块内类型

4. `practice / contest`
   - submit flag request 收回各自 `api/http`
   - submission response / record 收回各自 `application/commands`

5. `assessment report`
   - report export data 收回 `assessment/application/commands`
   - `assessment/api/http`、runtime wiring、app tests 改为消费模块内类型

6. 清理全局 DTO
   - 删除 `internal/dto/teacher.go`
   - 删除 `internal/dto/timeline.go`
   - 删除 `internal/dto/teacher_awd_review.go`
   - 删除 `internal/dto/submission.go`
   - 删除 `internal/dto/report.go`

## Expected Changes

- `code/backend/internal/module/teaching_query/**`
- `code/backend/internal/module/runtime/**`
- `code/backend/internal/module/assessment/**`
- `code/backend/internal/module/practice/**`
- `code/backend/internal/module/contest/**`
- `code/backend/internal/app/**test.go`
- `code/backend/cmd/seed-teaching-review-data/*.go`
- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/dto/timeline.go`
- `code/backend/internal/dto/teacher_awd_review.go`
- `code/backend/internal/dto/submission.go`
- `code/backend/internal/dto/report.go`

## Validation

- `go test ./internal/module/teaching_query/... -count=1`
- `go test ./internal/module/runtime/... -count=1`
- `go test ./internal/module/assessment/... -count=1`
- `go test ./internal/module/practice/... -count=1`
- `go test ./internal/module/contest/... -count=1`
- `go test ./internal/app -count=1`

## Review Focus

- owner 是否明确：`teaching_query`、`runtime`、`assessment`、`practice`、`contest` 各自只持有自己的 DTO
- query / handler / command 是否没有重新把全局 DTO 漏回 application 边界
- 删除全局教师侧 / submission / report DTO 后，seed / app test / export 链是否仍然正确
