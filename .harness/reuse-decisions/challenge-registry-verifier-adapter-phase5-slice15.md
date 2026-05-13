# Reuse Decision

## Change type

service / port / infrastructure / runtime

## Existing code searched

- `code/backend/internal/module/challenge/application/commands/registry_client.go`
- `code/backend/internal/module/challenge/application/commands/registry_client_test.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/infrastructure/`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/challenge/application/commands/registry_client.go`
- `code/backend/internal/module/challenge/infrastructure/image_repository.go`
- `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少新的 registry 校验能力，而是 `challenge/application/commands/registry_client.go` 本质上已经是一个实现 `challengeports.RegistryVerifier` 的 HTTP adapter，却还留在 application 包里直接依赖 `net/http`。最小正确方案不是新建一层更宽的 registry service，也不是把 image build service 再拆大，而是把现有 `RegistryClient` 原样下沉到 `challenge/infrastructure`，继续由 runtime 装配给 image build service 使用。

## Files to modify

- `code/backend/internal/module/challenge/application/commands/registry_client.go`
- `code/backend/internal/module/challenge/application/commands/registry_client_test.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-challenge-registry-verifier-adapter-phase5-slice15-implementation-plan.md`

## After implementation

- 如果后续继续收口 `auth` 的 CAS HTTP 校验，优先沿用这次“把外部 HTTP 调用 adapter 留在模块 infrastructure，由 runtime 装配给 application”的模式
- 如果 future challenge image build 还要增加 registry 变体，优先扩展 `challengeports.RegistryVerifier` 或新增同类 adapter，而不是把 HTTP 细节重新抬回 application
