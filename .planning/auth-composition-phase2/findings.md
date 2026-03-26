# Findings

- `auth` composition 仍直接读取 `identity.users` 私有字段，并在 `BuildAuthModule` 内把来自 `identity/ops` 的模块依赖直接摊开使用。
- `auth` 实际只需要 `UserRepository`、`TokenService`、profile command/query service、audit recorder 这几类稳定 contract。
- 这一轮不需要改业务实现，只要把 composition 收口到 typed deps，并补结构守卫防回退。
