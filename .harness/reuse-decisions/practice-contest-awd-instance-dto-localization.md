# Reuse Decision

## Change type

contract / api / mapper / practice / awd

## Existing code searched

- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/module/practice/api/http/*.go`
- `code/backend/internal/module/practice/application/commands/*.go`
- `code/backend/internal/app/full_router_integration_test.go`

## Similar implementations found

- `practice/api/http/submission_types.go` 已承接 practice HTTP request DTO
- `practice/application/commands/submission_output.go` 已承接 practice command output DTO
- `instance/contracts` / `runtime/api/http` 刚完成按 owner 收口的上一刀

## Decision

refactor_existing

## Reason

`internal/dto/contest_awd_instance.go` 现在只剩 practice admin AWD handler、command service 和对应 app 测试在用。继续留在全局 `dto` 只会让 `practice` 的请求类型、输出类型和 app 测试都跨模块依赖同一个桶。最小正确方案是把 request DTO 收回 `practice/api/http`，把 command output DTO 收回 `practice/application/commands`，再删除全局文件。

## Files to modify

- `.harness/reuse-decisions/practice-contest-awd-instance-dto-localization.md`
- `docs/plan/impl-plan/2026-05-18-practice-contest-awd-instance-dto-localization-implementation-plan.md`
- `code/backend/internal/dto/contest_awd_instance.go`
- `code/backend/internal/module/practice/api/http/contest_awd_instance_types.go`
- `code/backend/internal/module/practice/api/http/handler.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_instance_output.go`
- `code/backend/internal/module/practice/application/commands/contest_awd_operations.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service.go`
- `code/backend/internal/module/practice/application/commands/awd_scope_control_commands.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter.go`
- `code/backend/internal/module/practice/application/commands/response_mapper_goverter_gen.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/app/full_router_integration_test.go`

## After implementation

- `practice` admin AWD handler / commands / tests 不再引用 `dto.AdminAWD*` 或 `dto.*AdminContestAWD*`
- `internal/dto/contest_awd_instance.go` 删除
