# Reuse Decision

## Change type
service / composition / port / runtime

## Existing code searched
- `code/backend/internal/app/composition/`
- `code/backend/internal/module/practice/runtime/`
- `code/backend/internal/module/practice/application/commands/`
- `code/backend/internal/module/practice/contracts/`
- `code/backend/internal/module/assessment/runtime/`
- `code/backend/internal/module/assessment/application/commands/`
- `code/backend/internal/module/assessment/application/queries/`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Similar implementations found
- `code/backend/internal/module/practice/application/commands/submission_service.go`
- `code/backend/internal/module/practice/application/commands/manual_review_service.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`

## Decision
reuse_existing

## Reason
这次不需要新增 practice 到 assessment 的桥接 service、事件或 adapter。现有实现里，`practice` 已经在正确提交和人工评审通过时发布 `practice.flag_accepted`，`assessment` 也已经注册了 practice 事件消费者来做画像增量更新与推荐缓存刷新。剩余结构债只是 `practice` 仍保留一条历史同步直调链。最小正确方案是复用现有事件 contract 和消费者，删除同步注入与双轨实现，而不是再造新的 transition layer。

## Files to modify
- `code/backend/internal/app/composition/practice_module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/app/full_router_integration_test.go`
- `code/backend/internal/app/practice_flow_integration_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- `code/backend/internal/module/practice/application/commands/instance_start_service_test.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/practice/application/commands/service_audit_test.go`
- `code/backend/internal/module/practice/application/commands/service_test.go`
- `code/backend/internal/module/practice/application/commands/service_lifecycle_test.go`
- `code/backend/internal/module/practice/application/commands/submission_manual_review_test.go`
- `code/backend/internal/module/practice/architecture_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/design/backend-module-boundary-target.md`
- `docs/plan/impl-plan/2026-05-12-practice-assessment-boundary-phase3-slice1-implementation-plan.md`

## After implementation
- 如果这次收口形成稳定模式，再考虑把“已有事件消费者时，删除历史同步 service fallback”补进 `harness/reuse/history.md` 或更上层 skill。
