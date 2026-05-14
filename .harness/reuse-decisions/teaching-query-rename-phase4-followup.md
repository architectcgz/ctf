# Reuse Decision

## Change type

service / handler / repository / port / mapper / composition / api / readmodel

## Existing code searched

- `code/backend/internal/module/teaching_readmodel/`
- `code/backend/internal/module/teaching_query/`
- `code/backend/internal/module/practice/application/queries/`
- `code/backend/internal/module/assessment/contracts/`
- `code/backend/internal/app/composition/`
- `code/backend/internal/app/router*.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/practice/application/queries/`
- `code/backend/internal/module/assessment/contracts/`
- `code/backend/internal/module/runtime/runtime/module.go`
- `code/backend/internal/app/composition/instance_module.go`

## Decision

refactor_existing

## Reason

这次不是新增教师查询能力，也不是拆出新的读写分离层，而是复用已经存在的教师侧跨 owner 查询聚合实现，把误导性的 `teaching_readmodel` 命名统一收口为 `teaching_query`。

最小正确做法是：

- 保留现有 handler / query service / repository / ports 结构和行为不变
- 统一模块目录、composition 装配、router 依赖、测试守卫和 mapper 内部命名
- 同步当前事实文档，把“读模型”口径改成“查询聚合模块”，只在历史说明里保留 `practice_readmodel`

这样可以避免继续把教师查询模块描述成物理读写分离实现，同时不引入新的模块边界或重复实现。

## Files to modify

- `.harness/reuse-decisions/teaching-query-rename-phase4-followup.md`
- `docs/plan/impl-plan/2026-05-14-teaching-query-rename-implementation-plan.md`
- `README.md`
- `code/backend/internal/app/composition/teaching_query_module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/teaching_query/api/http/handler.go`
- `code/backend/internal/module/teaching_query/api/http/handler_contract_test.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
- `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
- `code/backend/internal/module/teaching_query/application/queries/contracts.go`
- `code/backend/internal/module/teaching_query/application/queries/overview_service.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper_assign.go`
- `code/backend/internal/module/teaching_query/application/queries/response_mapper_gen.go`
- `code/backend/internal/module/teaching_query/application/queries/service.go`
- `code/backend/internal/module/teaching_query/application/queries/service_overview_test.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
- `code/backend/internal/module/teaching_query/application/queries/student_review_service_test.go`
- `code/backend/internal/module/teaching_query/architecture_test.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/ports/query.go`
- `code/backend/internal/module/teaching_query/runtime/module.go`
- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/docs/reviews/backend/modular-monolith-refactor-checklist.md`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/02-database-design.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/architecture/features/攻击会话读模型与复盘工作台架构.md`
- `docs/architecture/features/攻击证据链与教学复盘架构.md`
- `docs/architecture/features/教学复盘优化设计.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/architecture/features/教师教学概览聚合架构.md`
- `docs/architecture/features/校园级CTF-AWD模式完整设计.md`
- `docs/design/backend-module-boundary-target.md`
- `docs/design/教学复盘建议优化方案.md`
- `harness/policies/project-patterns.yaml`

## After implementation

- `teaching_query` 继续表示教师侧跨 owner 查询聚合，而不是物理读写分离
- 当前事实文档统一使用“业务 owner / 查询聚合模块”口径
- 如果未来确实新增新的跨 owner 查询入口，优先复用 `teaching_query` 现有分层和测试守卫，而不是恢复 `readmodel` 命名
