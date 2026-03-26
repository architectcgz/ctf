# Findings

- `auth` 虽然已改为 `authModuleDeps`，但依然通过 `identity.users` 读取 `IdentityModule` 的私有仓储字段。
- `IdentityModule` 对外已经暴露了 profile/token contract，继续保留私有 `users` 作为跨模块输入会让边界表达不一致。
- 这轮只需要把用户仓储提升为公开 contract 字段 `Users`，并让 `auth` 改为依赖公开字段。
