# Reuse Decision

## Change type

runtime http / instance contract / query service

## Existing code searched

- `code/backend/internal/module/runtime/api/http/*`
- `code/backend/internal/module/instance/contracts/*`
- `code/backend/internal/module/instance/application/queries/instance_service.go`
- `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- `code/backend/internal/dto/teacher.go`

## Similar implementations found

- `identity/admin_user` 已经采用 `contracts + api/http` 的 owner 分层：跨模块服务边界走 `contracts`，HTTP request/response 由 `api/http` 自己持有
- `assessment teacher_awd_review` 响应 DTO 已收口到 owning module，并通过本地 mapper 保持外部 JSON 不变

## Decision

refactor_existing

## Reason

`TeacherInstanceQuery` 带 `form` / `binding` tag，属于 `runtime/api/http` 的 HTTP 输入；`TeacherInstanceItem` 则通过 `instance/contracts.InstanceQueryService` 跨模块暴露给 runtime facade，再由 handler 输出给外部。最小正确方案是把 query / item 从全局 `internal/dto/teacher.go` 拆成：

- `runtime/api/http` 持有 HTTP request / response DTO
- `instance/contracts` 持有跨模块查询契约

这样既能去掉全局 DTO 依赖，也不需要把 runtime 直接绑到 instance application 的具体实现。

## Files to modify

- `.harness/reuse-decisions/runtime-teacher-instance-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-runtime-teacher-instance-dto-localization-implementation-plan.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/reviews/backend/2026-05-17-runtime-teacher-instance-dto-localization-review.md`
- `code/backend/internal/module/instance/contracts/teacher_instance.go`
- `code/backend/internal/module/instance/contracts/services.go`
- `code/backend/internal/module/instance/application/queries/instance_service.go`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/runtime/api/http/handler_test.go`
- `code/backend/internal/module/runtime/api/http/teacher_instance_types.go`
- `code/backend/internal/module/runtime/api/http/teacher_instance_mapper.go`
- `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- `code/backend/internal/testutil/runtimeadapters/http_service.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/module/runtime/application/instance_service_test.go`
- `code/backend/internal/module/runtime/application/queries/response_mapper.go`
- `code/backend/internal/module/runtime/application/queries/response_mapper_gen.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/teacher.go`
