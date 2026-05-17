# Reuse Decision

## Change type

api / mapper / handler

## Existing code searched

- `code/backend/internal/dto/recommendation.go`
- `code/backend/internal/module/assessment/api/http/*.go`
- `code/backend/internal/module/teaching_query/application/queries/*.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `docs/architecture/backend/04-api-design.md`

## Similar implementations found

- `assessment/api/http/response_types.go` 已经承接学生端 recommendation response DTO
- `assessment/api/http/response_mapper.go` 已经改为输出模块内 response 类型
- `teaching_query` 教师端 recommendation DTO 仍在自己的消费边界，不依赖全局 `dto/recommendation.go`

## Decision

refactor_existing

## Reason

这次不是新增新的 response shape，也不是把 recommendation 再搬回全局共享桶，而是在上一刀 HTTP DTO 模块内化完成后，删除已经没有任何引用的旧全局文件：

- 学生端 recommendation HTTP DTO 已经由 `assessment/api/http` 持有
- 教师端 recommendation DTO 继续由 `teaching_query` 持有
- 全局 `internal/dto/recommendation.go` 已经失去 owner 和调用点

继续保留这个文件只会制造“全局 DTO 仍是有效事实源”的假象，所以最小正确动作就是删除它。

## Files to modify

- `.harness/reuse-decisions/assessment-recommendation-global-dto-removal.md`
- `docs/plan/impl-plan/2026-05-17-assessment-recommendation-global-dto-removal-implementation-plan.md`
- `docs/architecture/backend/04-api-design.md`
- `code/backend/internal/dto/recommendation.go`

## After implementation

- 全局 `internal/dto/recommendation.go` 被移除
- recommendation 的 HTTP DTO owner 只保留在模块边界内
- 后续若继续清理其他 DTO，可按同样方式逐条判断是否已有模块内 owner，再决定删除
