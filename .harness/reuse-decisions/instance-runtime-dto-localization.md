# Reuse Decision

## Change type

contract / api / mapper / runtime / instance

## Existing code searched

- `code/backend/internal/dto/instance.go`
- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/module/instance/**/*.go`
- `code/backend/internal/module/runtime/**/*.go`
- `code/backend/internal/module/practice/**/*instance*.go`
- `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- `code/backend/internal/testutil/runtimeadapters/http_service.go`

## Similar implementations found

- `teaching_query` 已把 query response DTO 收回 `contracts` / `application/queries`
- `assessment` 已把 report output 收回 `application/commands`
- `practice` 已把 submission request / output 收回 `api/http` / `application/commands`

## Decision

refactor_existing

## Reason

`internal/dto/instance.go` 混合了三类 owner：实例 owner 对外返回值、runtime HTTP 响应、AWD defense workbench 契约。继续挂在全局 `dto` 只会把 `instance / runtime / practice / app` 都绑在同一个桶上。最小正确方案是把实例 owner 稳定输出收回 `instance/contracts`，把仅 runtime HTTP 使用的响应收回 `runtime/api/http`，其余消费方改为依赖新 owner。

## Files to modify

- `.harness/reuse-decisions/instance-runtime-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-18-instance-runtime-dto-localization-implementation-plan.md`
- `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/dto/instance.go`
- `code/backend/internal/module/instance/application/awd_defense_workbench_service.go`
- `code/backend/internal/module/instance/application/awd_defense_workbench_service_test.go`
- `code/backend/internal/module/instance/application/commands/instance_service.go`
- `code/backend/internal/module/instance/application/queries/instance_service.go`
- `code/backend/internal/module/instance/contracts/instance_output.go`
- `code/backend/internal/module/instance/contracts/services.go`
- `code/backend/internal/module/instance/contracts/teacher_instance.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/practice/domain/mappers.go`
- `code/backend/internal/module/practice/domain/response_mapper_goverter.go`
- `code/backend/internal/module/practice/domain/response_mapper_goverter_gen.go`
- `code/backend/internal/module/runtime/api/http/access_response_types.go`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/runtime/api/http/handler_test.go`
- `code/backend/internal/module/runtime/api/http/teacher_instance_types.go`
- `code/backend/internal/module/runtime/application/commands/response_mapper.go`
- `code/backend/internal/module/runtime/application/commands/response_mapper_gen.go`
- `code/backend/internal/module/runtime/application/queries/response_mapper.go`
- `code/backend/internal/module/runtime/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/runtime/service_test.go`
- `code/backend/internal/testutil/runtimeadapters/http_service.go`

## After implementation

- 如果 `contest_awd_instance.go` 仍保留在 `internal/dto`，只允许它依赖模块 owner 类型，不再定义实例 owner 类型本身。
