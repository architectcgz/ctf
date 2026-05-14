# Reuse Decision

## Change type

service / repository / port / composition / readmodel

## Existing code searched

- `code/backend/internal/app/composition/identity_module.go`
- `code/backend/internal/app/composition/teaching_query_module.go`
- `code/backend/internal/module/identity/contracts/auth.go`
- `code/backend/internal/module/identity/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/runtime/module.go`
- `code/backend/internal/module/teaching_query/application/queries/*.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`

## Similar implementations found

- `code/backend/internal/module/auth/application/commands/service.go`
- `code/backend/internal/module/auth/application/commands/cas_service.go`
- `code/backend/internal/app/composition/auth_module.go`

## Decision

refactor_existing

## Reason

这次不是把教师目录和复盘聚合改成通过 `identity` 做二次拼装，也不是新增新的用户查询模块，而是复用已经存在的 `identity.Users` 基础 lookup 能力，只替换 `teaching_query` 里重复实现的 `FindUserByID` 路径。

最小正确方案是：

- `teaching_query` 继续保留班级、学生目录、活跃度、得分、薄弱维度等跨表聚合 SQL
- `FindUserByID` 这类基础用户 lookup 改为由 composition 注入 `identity.Users`
- `teaching_query` 内部 query service 改成“用户 lookup + 聚合 repo”分离，而不是继续让本地 repo 同时承担两种职责
- 同步架构文档，把 `teaching_query` 的依赖口径从“只依赖 assessment”更新为“assessment + identity lookup（装配）”

这样可以复用现有 owner contract，避免把教师端聚合查询拆成失真的二次分页/二次排序。

## Files to modify

- `.harness/reuse-decisions/teaching-query-identity-user-lookup.md`
- `docs/plan/impl-plan/2026-05-14-teaching-query-identity-user-lookup-implementation-plan.md`
- `code/backend/internal/app/composition/teaching_query_module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/module/teaching_query/application/queries/service.go`
- `code/backend/internal/module/teaching_query/application/queries/overview_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/application/queries/service_overview_test.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service_test.go`
- `code/backend/internal/module/teaching_query/runtime/module.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/ports/query.go`
- `code/backend/internal/module/teaching_query/architecture_test.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`

## After implementation

- `teaching_query` 不再在本地 repo 里重复承担基础用户 lookup
- 教师侧聚合查询仍保持单次 SQL 聚合，不回退成 `identity list + teaching_query` 二次拼装
- 如果未来还需要复用更丰富的用户 roster 能力，再考虑从 `identity` 暴露更窄的教师查询 contract，而不是恢复本地重复 lookup
