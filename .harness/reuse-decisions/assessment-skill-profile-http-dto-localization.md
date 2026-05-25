# Reuse Decision

## Change type
handler / api / mapper / service

## Existing code searched
- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/module/assessment/application/queries/profile_service.go`
- `code/backend/internal/module/assessment/application/commands/profile_service.go`
- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/domain/profile.go`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`

## Similar implementations found
- `assessment` recommendation 已经把 query contract 放在 `contracts`，再由 `assessment/api/http` 持有本地 response DTO 与 mapper。
- `assessment/api/http/response_types.go` 和 `response_mapper.go` 已经是 recommendation HTTP DTO 本地化的既有落点。
- `assessment/runtime/module.go` 已经把 query service 和 report service 分开装配，允许 skill profile 查询链与 report 读取链分别走不同边界。

## Decision
extend_existing

## Reason
这次不是新增一套 skill profile 能力，而是沿用 recommendation 已经落地的“模块内 contract + HTTP 本地 DTO”模式，把 skill profile 的 query / HTTP owner 从全局 `internal/dto` 收回 `assessment` 模块。

最小改动面是：

- 在 `assessment/contracts` 增加 skill profile contract 类型，供 query service 输出。
- 在 `assessment/api/http` 增加本地 response DTO 与 mapper，保持 JSON 契约不变。
- 复用 `runtime/module.go` 现有分层装配，让 handler 走 query contract，而 report 继续使用现有 profile reader，避免把本刀扩成 report 重构。

这样可以完成本次目标，同时把 `internal/dto/skill_profile.go` 的删除判断留给实现后的实际引用结果，而不是先扩大 touched surface。

## Files to modify
- `.harness/reuse-decisions/assessment-skill-profile-http-dto-localization.md`
- `docs/plan/archive/impl-plan/2026-05-17-assessment-skill-profile-http-dto-localization-implementation-plan.md`
- `code/backend/internal/module/assessment/contracts/contracts.go`
- `code/backend/internal/module/assessment/api/http/handler.go`
- `code/backend/internal/module/assessment/api/http/response_types.go`
- `code/backend/internal/module/assessment/api/http/response_mapper.go`
- `code/backend/internal/module/assessment/api/http/response_mapper_gen.go`
- `code/backend/internal/module/assessment/application/queries/profile_service.go`
- `code/backend/internal/module/assessment/runtime/module.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/module/assessment/application/queries/profile_service_test.go`

## After implementation
- 如果后续 report 也切到模块内 contract，再评估是否可以删除 `code/backend/internal/dto/skill_profile.go`。
- 若这条模式后续会重复用于 assessment 其他查询边界，再补 `harness/reuse/index.yaml`；本次先保留 task-scoped evidence。
