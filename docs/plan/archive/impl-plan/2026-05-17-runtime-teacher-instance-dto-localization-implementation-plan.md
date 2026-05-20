# runtime teacher instance DTO 模块内化实现方案

## Objective

把教师实例列表的 query / item 从全局 `internal/dto/teacher.go` 收回 owning module，明确 `runtime/api/http` 与 `instance/contracts` 的边界，同时保持 `GET /api/v1/teacher/instances` 的外部查询参数和 JSON 字段不变。

## Non-goals

- 不改学生侧实例接口：`/api/v1/instances/*`
- 不改实例列表的过滤语义、教师越权校验和 access URL 生成逻辑
- 不扩到 `teacher.go` 里的 class / overview / evidence / attack session 这些 teaching query 聚合 DTO

## Inputs

- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-17-teacher-aggregate-dto-localization-implementation-plan.md`
- `code/backend/internal/module/runtime/api/http/handler.go`
- `code/backend/internal/module/instance/contracts/services.go`
- `code/backend/internal/module/instance/application/queries/instance_service.go`
- `code/backend/internal/app/composition/runtime_http_service_adapter.go`
- `code/backend/internal/dto/teacher.go`

## Task Slices

1. 新增 runtime / instance 本地类型
   - `runtime/api/http`：教师实例列表 query / response DTO
   - `instance/contracts`：教师实例列表 query / item contract

2. 收口主链路接口
   - `runtime/api/http` handler 改为绑定本地 query，再映射到 `instance/contracts`
   - `app/composition/runtime_http_service_adapter.go` 与 `instance/contracts.InstanceQueryService` 改为使用模块内 contract
   - `instance/application/queries/instance_service.go` 改为返回 contract item

3. 清理测试与 compat mapper
   - runtime 相关测试和 full router decode 改为使用新 owner
   - `runtime/application/queries/response_mapper*` 去掉已无 owner 的 `TeacherInstanceItem` 生成映射
   - 从 `internal/dto/teacher.go` 删除 `TeacherInstanceQuery` / `TeacherInstanceItem`

## Expected Changes

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

## Validation

- `go generate ./internal/module/runtime/application/queries`
- `go test ./internal/module/runtime/api/http -count=1`
- `go test ./internal/module/runtime -run 'TestServiceListTeacherInstances' -count=1`
- `go test ./internal/module/runtime/application -run 'TestInstanceServiceListTeacherInstances|TestInstanceServiceListTeacherInstancesPrefersContestAWDServiceMetadata|TestInstanceServiceListTeacherInstancesNormalizesTCPAccessURL' -count=1`
- `go test ./internal/app -run 'TestFullRouter_InstanceHintAndProxyStateMatrix' -count=1`

## Review Focus

- `TeacherInstanceQuery` 是否已经只由 `runtime/api/http` 持有
- `TeacherInstanceItem` 是否已经从全局 DTO 收口到 `instance/contracts`，并由 runtime handler 显式映射为 HTTP DTO
- 删除全局 `teacher.go` 中这两个类型后，runtime / instance / composition / app 测试链是否仍然保持既有行为
