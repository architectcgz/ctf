# Findings

- `runtime` 已完成对外 bridge 收口，但 `BuildRuntimeModule` 仍把 repo、engine、application service、cross-module adapter、后台任务注册混在一个函数里。
- 当前 `RuntimeModule` 内部依赖已经稳定，可以拆为 typed deps 与按能力分组的局部 builder，而不改变任何对外 contract。
- 这轮重点是 composition 标准化，不触碰 runtime adapter 语义和外部路由。
