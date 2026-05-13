# Reuse Decision

## Change type

service / port / infrastructure / runtime

## Existing code searched

- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/auth/infrastructure/cas_ticket_validator.go`
- `docs/design/backend-module-boundary-target.md`

## Similar implementations found

- `code/backend/internal/module/auth/infrastructure/cas_ticket_validator.go`
- `code/backend/internal/module/challenge/infrastructure/registry_client.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`

## Decision

refactor_existing

## Reason

当前 debt 不是缺少实例探活能力，而是 `practice/application/commands/instance_provisioning.go` 同时持有 provisioning 编排、重试窗口和 HTTP/TCP 探活细节。最小正确方案不是把整个 provisioning service 拆散，也不是把 runtime service 再做宽，而是沿用 phase5 已验证模式：application 继续保留“什么时候探活、重试几次、失败如何标记实例失败”的编排语义，新增模块内窄 `PracticeInstanceReadinessProbe` port，由 infrastructure 统一承接 HTTP/TCP access URL 探测细节。

## Files to modify

- `code/backend/internal/module/practice/application/commands/service.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning.go`
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
- `code/backend/internal/module/practice/application/commands/contest_instance_service_test.go`
- `code/backend/internal/module/practice/ports/ports.go`
- `code/backend/internal/module/practice/ports/instance_readiness_probe_context_contract_test.go`
- `code/backend/internal/module/practice/infrastructure/instance_readiness_probe.go`
- `code/backend/internal/module/practice/infrastructure/instance_readiness_probe_test.go`
- `code/backend/internal/module/practice/runtime/module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/plan/impl-plan/2026-05-13-practice-instance-readiness-probe-phase5-slice17-implementation-plan.md`

## After implementation

- 如果后续 `contest` 的 AWD checker / probe 继续收口 HTTP concrete，优先复用这次“application 保留重试编排，infrastructure 承接协议探测”的模式
- 如果未来实例探活还要支持新协议，优先扩 `PracticeInstanceReadinessProbe` 的 infrastructure 实现，不把网络细节重新抬回 application
