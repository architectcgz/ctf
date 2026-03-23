# Findings

- 当前 `runtimeinfra` 只在 composition 和 `runtime` 内部桥接中被使用。
- `runtimeinfra_module.go` 只是为 `Engine` 创建单独模块，已经明显是过渡性结构。
- 架构规则仍然专门放行 `runtimeinfra` 根包，说明这条迁移尚未结束。
- `Engine` 的受管容器查询能力可以直接下沉到 `runtime/infrastructure`，不必继续通过 composition 里的额外 adapter 转一层。
