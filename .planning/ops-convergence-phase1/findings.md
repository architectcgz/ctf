# Findings

- `ops` 迁移不能一口气把 `notification` 一并搬走，因为它还耦合 websocket 握手、ticket 校验与事件消费。
- `audit / dashboard / risk` 是可独立先迁的第一批 owner，且对外 HTTP 路径可以保持不变。
- `composition.SystemModule` 可以继续作为 app 侧聚合点，但其审计/仪表盘/风控字段应切到 `ops` contract，而不是继续暴露 `system` concrete。
- 仪表盘仍需要 runtime 运行指标，但依赖可以通过 query / stats provider bridge 收窄，不必让 `ops` 直接依赖 runtime persistence concrete。
