# Reuse Decision

## Change type

handler / service / cache / contract / api

## Existing code searched

- `code/backend/internal/module/practice/api/http/*.go`
- `code/backend/internal/module/practice/api/http/progress_dto.go`
- `code/backend/internal/module/practice/api/http/response_mapper.go`
- `code/backend/internal/module/practice/api/http/response_mapper_assign.go`
- `code/backend/internal/module/practice/api/http/response_mapper_gen.go`
- `code/backend/internal/module/practice/application/queries/*.go`
- `code/backend/internal/module/practice/ports/*.go`
- `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- `code/backend/internal/dto/progress.go`
- `code/backend/internal/dto/*.go`
- `code/backend/internal/app/*practice*test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Similar implementations found

- `assessment/api/http` 已经把学生端 recommendation HTTP DTO 收回模块边界，说明 HTTP DTO owner 适合落在 `api/http`
- `practice/ports` 已经承接 repository record、cache contract 等模块内查询契约，适合继续承接 progress cache 的内部快照类型
- `internal/dto` 仍然被 `teaching_query` 使用为教师端 timeline DTO 共享桶，因此本刀不能把 teaching query 一并迁走

## Decision

refactor_existing

## Reason

这次不新增新的 shared DTO 包，也不把 `practice/application/queries` 反向耦合到 `api/http`。最小可审查做法是：

- `practice/api/http` 新增模块内 progress/timeline HTTP DTO，收回学生端 practice 接口的响应 owner
- `practice/ports` 新增 progress cache / query 使用的模块内快照类型，避免继续依赖全局 `internal/dto`
- `practice/application/queries` 在模块内快照与 HTTP DTO 之间保持简单映射，不把 teaching query 拉进本次切片
- `internal/dto/progress.go` 若仅剩 timeline 共享类型，则把 timeline 类型迁到新的 `internal/dto/timeline.go` 后删除 `progress.go`

这样复用了现有 practice 分层、handler/query/cache 结构和 app 集成测试入口，只收口当前切片真正触达的 owner。

## Files to modify

- `.harness/reuse-decisions/practice-progress-timeline-http-dto-localization.md`
- `docs/plan/impl-plan/2026-05-17-practice-progress-timeline-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/practice/api/http/*.go`
- `code/backend/internal/module/practice/api/http/progress_dto.go`
- `code/backend/internal/module/practice/api/http/response_mapper.go`
- `code/backend/internal/module/practice/api/http/response_mapper_assign.go`
- `code/backend/internal/module/practice/api/http/response_mapper_gen.go`
- `code/backend/internal/module/practice/application/queries/*.go`
- `code/backend/internal/module/practice/ports/*.go`
- `code/backend/internal/module/practice/infrastructure/progress_cache.go`
- `code/backend/internal/app/practice_flow_integration_test.go`（如需）
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`（如需）
- `code/backend/internal/dto/progress.go`
- `code/backend/internal/dto/timeline.go`

## After implementation

- `practice` 的 progress/timeline HTTP 响应不再依赖 `internal/dto/progress.go`
- progress cache / query contract 使用 practice 模块内类型
- 是否删除 `internal/dto/progress.go` 以全仓引用结果为准
