# Runtime Teacher Instance DTO Localization Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 runtime teacher instance DTO 模块内化相关 diff，重点覆盖 `runtime/api/http`、`instance/contracts`、`instance/application/queries`、composition adapter 与教师实例列表相关测试
- Files reviewed:
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
  - `docs/architecture/backend/04-api-design.md`
  - `docs/plan/impl-plan/2026-05-17-runtime-teacher-instance-dto-localization-implementation-plan.md`
  - `.harness/reuse-decisions/runtime-teacher-instance-dto-localization.md`
- Classification check: agree with pipeline，属于 non-trivial backend refactor + review gate
- Initial gate verdict: pass with minor issues

## Current Status

- 2026-05-17 补充修复状态：已完成。独立 review 没有发现 blocker；补充项主要是教师实例列表的 HTTP 契约断言偏弱，当前已新增 `TestListTeacherInstancesBindsQueryIntoInstanceContract`，并加强 `TestFullRouter_InstanceHintAndProxyStateMatrix` 对 `student_username`、`access_url`、`remaining_time` 字段的断言。

## Findings

- no findings

## Validation Evidence

- `cd code/backend && go generate ./internal/module/runtime/application/queries`
- `cd code/backend && go test ./internal/module/runtime/api/http -count=1`
- `cd code/backend && go test ./internal/module/runtime -run 'TestServiceListTeacherInstances' -count=1`
- `cd code/backend && go test ./internal/module/runtime/application -run 'TestInstanceServiceListTeacherInstances|TestInstanceServiceListTeacherInstancesPrefersContestAWDServiceMetadata|TestInstanceServiceListTeacherInstancesNormalizesTCPAccessURL' -count=1`
- `cd code/backend && go test ./internal/module/runtime/application/... -count=1`
- `cd code/backend && go test ./internal/module/instance/contracts ./internal/module/instance/application/queries -count=1`
- `cd code/backend && go test ./internal/app -run 'TestFullRouter_InstanceHintAndProxyStateMatrix' -count=1`

## Final Review Verdict

- Gate verdict after follow-up tests: pass
- 结论：`GET /api/v1/teacher/instances` 的 request / response DTO 已不再依赖全局 `internal/dto/teacher.go`；HTTP 输入输出 owner 收口到 `runtime/api/http`，跨模块查询契约收口到 `instance/contracts`，教师实例列表的过滤语义、越权校验和 JSON 字段保持不变。
