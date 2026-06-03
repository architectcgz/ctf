# Reuse Decision

## Change type
e2e validation / script / composition integration test

## Existing code searched
- code/backend/internal/app/composition
- code/backend/internal/module/runtime/infrastructure/agentclient
- code/backend/internal/module/runtime/infrastructure/agentserver
- code/backend/internal/bootstrap
- scripts
- docs/operations/runtime-agent-deployment.md

## Similar implementations found
- `code/backend/internal/app/composition/runtime_node_execution_router_test.go` 已经覆盖了 cleanup authority 的单元级回归，但当前仍是 stub client，不能证明真实 agent + 多 Docker daemon 下的行为。
- `code/backend/internal/module/runtime/infrastructure/agentclient/bridge_integration_test.go` 已经覆盖 mTLS gRPC 基础桥接，说明本轮不需要重新设计 agent 协议，只需要把真实 agent 与真实 Docker executor 串起来。
- `code/backend/internal/bootstrap/runtime_agent.go` 已经提供独立 `runtime-agent` 进程入口，适合被脚本直接拉起，不需要额外做测试专用二进制。

## Decision
extend_existing

## Reason
这次用户要验证的是“本地能否模拟真双节点，并通过脚本复现实例实际跑在 node B、cleanup payload 却缺少 NodeID 时，router 仍能清到 node B”。

最小正确路径不是再去扩充 compose，也不是手工写一串临时命令让用户自己拼，而是：

- 保留现有单元测试继续约束 owner
- 新增一条受环境变量驱动的 composition e2e 测试，真实连两台 `runtime-agent`
- 用脚本自动起两套独立 `dockerd`、两台 agent 和临时证书，再调用该测试

这样能证明的不是“node 表里有两条记录”，而是更关键的：

- 两台 agent 背后是两套不同 Docker daemon
- 默认节点与实例真实宿主可以分离
- 缺失 `NodeID` 的 cleanup 会通过 `container_id` / `runtime_details` 反查，最终落到正确宿主

## Files to modify
- .harness/reuse-decisions/runtime-control-plane-agent-split-dual-node-e2e-validation.md
- code/backend/internal/app/composition/runtime_node_execution_router_e2e_test.go
- scripts/runtime-agent-dual-node-e2e.sh

## After implementation
- 仓库会有一条可重复执行的本地真双节点验证路径，而不是只剩口头步骤。
- cleanup authority 的 correctness 会同时有单元测试和真实 agent / 双 Docker daemon 的 e2e 证据。
