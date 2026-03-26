# Findings

- [`practice_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/practice_module.go) 仍保留大量 runtime bridge 代码：`practiceRuntime*Bridge`、`practiceRuntimeInstanceServiceAdapter`、以及 `runtimeapp` 请求/响应映射函数。
- 这些 bridge 实际是为了把 `runtime/application` 的 provisioning 与 cleanup 能力适配给 `practice` 使用，更合理的位置应是 [`runtime_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/runtime_module.go)。
- [`runtime_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/runtime_module.go) 还直接使用 `contestinfra.AWDContainerFileWriter` 类型，导致 runtime composition 反向依赖 `contest/infrastructure`。
- 下一刀适合收口为“runtime 向外暴露外部模块可用的接口依赖”，让 `practice` 只消费 `practiceports.*`，`contest` 只消费 `contestports.*`。
