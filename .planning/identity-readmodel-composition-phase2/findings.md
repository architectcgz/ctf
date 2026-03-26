# Findings

- `identity_module.go`、`practice_readmodel_module.go`、`teaching_readmodel_module.go` 仍是较早期的 composition 写法，直接在 `Build*Module` 函数里 new concrete repository/service。
- 这些模块的业务层和 ports/contracts 已经足够稳定，本轮不需要继续拆应用层接口，重点只是让 composition 依赖类型与其他模块风格对齐。
- 这组改动不涉及外部 API 或路由，只是把装配层标准化，适合作为轻量收口切片推进。
