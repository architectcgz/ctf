# Reuse Decision

## Change type
port / service / composition / runtime host execution boundary

## Existing code searched
- code/backend/internal/app/composition/runtime_module.go
- code/backend/internal/app/composition/contest_module.go
- code/backend/internal/module/runtime/ports/container_runtime.go
- code/backend/internal/module/runtime/infrastructure/engine.go
- code/backend/internal/module/runtime/infrastructure/engine_files.go
- code/backend/internal/module/runtime/infrastructure/engine_provisioning.go
- code/backend/internal/module/runtime/runtime/module.go
- code/backend/internal/module/contest/runtime/module.go
- code/backend/internal/module/contest/ports/checker_runner.go
- code/backend/internal/module/contest/infrastructure/docker_checker_runner.go
- code/backend/internal/app/composition/runtime_module_test.go
- code/backend/internal/app/router_test.go

## Similar implementations found
- `runtime/runtime/module.go` 已经把 challenge / ops / contest 侧 runtime 能力收口成 typed deps，说明新的执行面边界应该继续走显式 dep 注入，而不是在子模块内部直接 new Docker 细节。
- `composition/*_module.go` 当前负责 app root 装配，适合作为 local adapter 的唯一接入点；模块内部 runtime / contest 代码已经更多围绕 port 工作，而不是直接依赖 app root。
- `docker_checker_runner.go` 与 `engine*.go` 已经分别形成 checker sandbox 与 container host execution 的本地实现，可以先复用现有实现，只把 owner 和 wiring 抽清。

## Decision
refactor_existing

## Reason
这轮不直接进入 remote agent 协议实现，先把“本地宿主机执行”从隐含事实改成显式 local adapter：

- 在 `runtime/ports` 明确 runtime host execution 聚合边界
- 保留 `Engine` 作为本地 adapter 实现，但通过显式 constructor 暴露给 composition
- 把 AWD checker runner 的本地 Docker 构造从 `contest/runtime/module.go` 抽到 composition，模块内部只依赖注入的 checker execution capability

这样能在不改业务语义的前提下，为后续 remote agent adapter 预留稳定接入点，避免继续把 Docker client / checker sandbox 构造散在控制面模块里。

## Files to modify
- .harness/reuse-decisions/runtime-control-plane-agent-split-slice1.md
- code/backend/internal/module/runtime/ports/container_runtime.go
- code/backend/internal/module/runtime/infrastructure/engine.go
- code/backend/internal/module/contest/infrastructure/docker_checker_runner.go
- code/backend/internal/module/contest/runtime/module.go
- code/backend/internal/app/composition/runtime_module.go
- code/backend/internal/app/composition/contest_module.go
- code/backend/internal/app/composition/runtime_module_test.go
- code/backend/internal/app/router_test.go

## After implementation
- `local host execution` 与 `local checker execution` 会有显式 adapter 入口。
- contest / runtime 子模块不再在内部直接 new Docker checker runner。
- app composition 只负责装配 host execution capability，而不是继续散落低层 Docker 细节。
