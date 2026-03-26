# Findings

- `ops` 模块的 ports 已经完整，但 composition 仍保留早期写法：`BuildOpsModule` 与 `BuildNotificationHandler` 直接 inline new concrete repository/service。
- `ops` 的 dashboard 依赖当前通过 `runtime.ops.query / statsProvider` 读取 `RuntimeModule` 内部字段，边界比近期其他模块更松。
- 这轮适合只做 composition 标准化：引入 typed deps、拆局部 builder，并让 runtime 到 ops 的适配留在 runtime composition 内部。
