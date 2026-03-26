# Findings

- `practice` 已完成仓储端口收口，但 composition 仍把持久化依赖与 `challenge/runtime/assessment` 跨模块依赖混在同一个 builder 里。
- 当前跨模块输入已经稳定为 `InstanceRepository`、`RuntimeInstanceService`、challenge contract、image store、assessment profile service。
- 这轮适合只做 composition 标准化：拆出 external deps builder 和 handler builder，让模块边界表达与前几轮一致。
