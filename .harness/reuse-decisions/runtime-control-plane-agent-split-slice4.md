# Reuse Decision

## Change type
port / service / composition

## Existing code searched
- code/backend/internal/module/runtime/infrastructure
- code/backend/internal/module/runtime/infrastructure/agentclient
- code/backend/internal/module/runtime/infrastructure/agentserver
- code/backend/internal/module/runtime/application/commands
- code/backend/internal/app/composition

## Similar implementations found
- `code/backend/internal/module/runtime/infrastructure/agentclient/bridge.go` 已经实现 `ApplyACLRules`、`ApplyACL`、`RemoveACLRules`、`RemoveACL` 的远端桥接，说明 phase4 不需要再新增第二套 ACL RPC。
- `code/backend/internal/module/runtime/infrastructure/agentserver/service.go` 已经把 ACL RPC 透传到 `RuntimeHostExecutor`，宿主执行 owner 已经有明确落点。
- `code/backend/internal/module/runtime/application/commands/provisioning_service.go` 与 `runtime_cleanup_service.go` 已经通过 `ContainerProvisioningRuntime` / `ContainerCleanupRuntime` 抽象消费 ACL 能力，业务层本身不再拼接 `iptables` 命令。
- `code/backend/internal/module/runtime/service_test.go`、`service_acl_test.go` 已经覆盖 ACL handle 在 provisioning / cleanup 层的本地语义，缺口主要在远端 agent delegation 证据。

## Decision
extend_existing

## Reason
phase4 的目标是把 ACL authority 明确放到宿主执行侧，而不是在 API 侧重写 ACL 实现。当前代码里：

- composition 在 remote mode 下已经把 runtime executor 收口为 agent bridge
- bridge / agentserver 已经包含 ACL 调用面
- provisioning / cleanup 已经只依赖 runtime port，而不直接知道 `iptables` 细节

因此这轮最小正确路径是基于现有实现补齐 phase4 证据与缺失测试，确认：

- remote mode 下 `ApplyACL` / `RemoveACL` 确实通过 agent 调到宿主执行侧
- API 侧运行时编排继续只面向 port，不重新泄漏 `iptables` owner

如果补证据时发现调用链没有按预期经过 bridge，再回到实现层最小修复；否则不做无谓重构。

## Files to modify
- .harness/reuse-decisions/runtime-control-plane-agent-split-slice4.md
- code/backend/internal/module/runtime/infrastructure/agentclient/bridge_integration_test.go
- code/backend/internal/module/runtime/infrastructure/agentserver/service_test.go
- code/backend/internal/module/runtime/service_test.go

## After implementation
- phase4 会有独立的 ACL owner 迁移证据，不再依赖 phase2 “bridge 已经存在”的隐含推断。
- 如果最终确认 phase4 生产代码已在 phase2 完成，本轮只保留测试与 task evidence，不人为扩大 touched surface。
