# Reuse Decision

## Change type
port / composition / config / service / agent protocol / command

## Existing code searched
- code/backend/internal/module
- code/backend/internal/app/composition
- code/backend/internal/module/runtime/ports
- code/backend/internal/module/runtime/infrastructure
- code/backend/internal/module/contest/infrastructure
- code/backend/internal/config
- code/backend/cmd
- code/backend/configs

## Similar implementations found
- `code/backend/internal/app/composition/runtime_module.go` 已经把 runtime host execution 收口成 composition 注入点，适合继续在这里选择 local / remote adapter。
- `code/backend/internal/app/composition/contest_module.go` 与 `contest/runtime/module.go` 已经把 checker runner 从模块内部构造抽成显式依赖，说明 phase2 可以沿同一条 execution bridge 把 checker 也接到 remote agent。
- `code/backend/internal/bootstrap/run.go`、`cmd/api/main.go` 已经提供了 API 进程的启动/日志/优雅退出模式；新的 agent command 应尽量复用同样的启动结构，而不是临时拼接 shell。
- 当前仓库没有现成的 protobuf / buf 工作流，也没有现成的 gRPC service 定义生成链，说明这轮需要在现有 Go 模块内引入最小可维护的协议实现方式。

## Decision
refactor_existing

## Reason
phase2 的目标不是立即迁完 checker bind mount 或 ACL owner，而是先把“API 通过正式协议调用宿主执行面”做成真实可用的 adapter。最小正确落点是：

- 在 `runtime/agentcontracts` 定义显式协议与 mTLS 传输约束
- 在 `runtime/infrastructure/agentclient` 实现 remote host executor + checker runner
- 在 `cmd/runtime-agent` 落一个可启动的 agent server，内部复用现有 local engine / checker runner
- 在 composition / config 层增加单节点 remote agent 选择逻辑，同时保留 local fallback

这样 slice3 / slice4 可以继续沿同一条通道迁 checker 文件与 ACL owner，而不是再新增第二套远端执行路径。

## Files to modify
- .harness/reuse-decisions/runtime-control-plane-agent-split-slice2.md
- code/backend/internal/config/config.go
- code/backend/internal/config/config_test.go
- code/backend/configs/config.yaml
- code/backend/internal/app/composition/runtime_module.go
- code/backend/internal/app/composition/contest_module.go
- code/backend/internal/app/router.go
- code/backend/internal/app/router_test.go
- code/backend/internal/module/runtime/agentcontracts/**
- code/backend/internal/module/runtime/infrastructure/agentclient/**
- code/backend/internal/module/runtime/infrastructure/agentserver/**
- code/backend/internal/module/contest/infrastructure/docker_checker_runner.go
- code/backend/internal/module/runtime/infrastructure/engine.go
- code/backend/cmd/runtime-agent/**

## After implementation
- API 侧会有明确的 remote runtime agent 配置与 adapter。
- runtime-agent 会成为可独立启动的宿主执行进程骨架。
- local / remote execution 选择不再散落在业务模块内部，而是在 composition 与 config 层集中控制。
