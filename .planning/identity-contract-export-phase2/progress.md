# Progress

- 2026-03-26: 启动 `identity-contract-export-phase2`，目标是消除对 `IdentityModule` 私有仓储字段的跨模块访问。
- 2026-03-26: 已补结构守卫，要求 `IdentityModule` 公开暴露 `Users` contract，并禁止 `auth` 回退到 `identity.users`。
- 2026-03-26: `identity_module.go` 已将用户仓储提升为公开字段 `Users`，`auth_module.go` 已切换为读取公开 contract。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestIdentityModuleUsesTypedDeps|TestAuthModuleUsesTypedDeps|TestNewRouterRegistersStudentChallengeRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/identity/... ./internal/module/auth/... -count=1`
