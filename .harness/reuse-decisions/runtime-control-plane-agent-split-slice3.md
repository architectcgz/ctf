# Reuse Decision

## Change type
port / service / composition

## Existing code searched
- code/backend/internal/module/contest/infrastructure
- code/backend/internal/module/contest/ports
- code/backend/internal/module/runtime/infrastructure
- code/backend/internal/module/runtime/infrastructure/agentclient
- code/backend/internal/module/runtime/infrastructure/agentserver
- code/backend/internal/app/composition

## Similar implementations found
- `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go` 已经集中承担 checker sandbox 容器的创建、启动、日志采集与清理，phase3 继续在这里收口“文件如何进入容器”的 owner 最小且直接。
- `code/backend/internal/module/runtime/infrastructure/engine_files.go` 已经有通过 tar 流 + Docker `CopyToContainer` 写入容器文件的实现模式，说明 phase3 不需要新引入第二套文件注入协议。
- `code/backend/internal/module/runtime/infrastructure/agentclient/bridge.go` 与 `code/backend/internal/module/runtime/infrastructure/agentserver/service.go` 已经把 `contestports.CheckerRunJob` 原样通过 agent 协议传递，`CheckerRunJob.Files` 可直接承载 checker 文件内容。

## Decision
refactor_existing

## Reason
phase3 的目标不是改 checker job 协议，而是把 checker 文件 owner 从“API 本机临时目录 + bind mount”迁到“执行侧 Docker 文件注入”。当前仓库已经具备：

- `CheckerRunJob.Files` 作为文件载荷
- `DockerCheckerRunner` 作为唯一 checker sandbox 执行 owner
- runtime 侧现成的 `CopyToContainer` tar 注入模式

因此最小正确路径是重构现有 `DockerCheckerRunner`：

- 容器创建时不再依赖 host bind mount
- 创建后、启动前把 `job.Files` 通过 tar 流复制到容器工作目录
- 只把 `HostWorkRoot` 保留为 agent 宿主可选临时工作根语义，不再作为 API 本机路径前提

这样可以让 local mode 和 remote agent mode 共享同一条 checker 文件注入逻辑，并为后续 phase4 ACL owner 迁移保留同样的 execution bridge。

## Files to modify
- .harness/reuse-decisions/runtime-control-plane-agent-split-slice3.md
- code/backend/internal/app/composition/runtime_module.go
- code/backend/internal/app/composition/runtime_module_test.go
- code/backend/internal/config/config.go
- code/backend/internal/config/config_test.go
- code/backend/configs/config.yaml
- code/backend/internal/module/contest/infrastructure/docker_checker_runner.go
- code/backend/internal/module/contest/infrastructure/docker_checker_runner_test.go

## After implementation
- checker sandbox 执行路径不再要求 API 本机目录与 Docker 宿主同机。
- agent 侧 `RunChecker` 会直接消费 `CheckerRunJob.Files` 并在宿主执行侧注入容器文件。
- 如后续还有容器文件批量注入需求，再评估是否把 tar 归档 helper 提取成共享实现；这轮先保持 phase3 变更面聚焦在 checker runner。
