# Challenge Layering Phase2

- [ ] 为 `challenge` 补结构守卫测试，禁止继续声明 legacy 宽 `ChallengeRepository`，并要求 composition 使用 typed deps 收口。
- [ ] 按应用用例拆分 `challenge/ports` 中的主仓储接口，更新 commands/queries 构造依赖到窄接口。
- [ ] 调整 `challenge` composition 装配，让模块依赖通过 ports/contracts 暴露，避免 deps 持有具体 infra 仓储类型。
- [ ] 运行 `challenge` 模块测试与 app 侧定向 router/composition 测试，确认切片可独立回归。
