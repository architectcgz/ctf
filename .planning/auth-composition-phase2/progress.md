# Progress

- 2026-03-26: 启动 `auth-composition-phase2`，目标是收紧 `auth` composition 对 `identity/ops` 的依赖边界。
- 2026-03-26: 已补结构守卫，要求 `BuildAuthModule` 使用 `authModuleDeps`，不再直接把 `identity/ops` 模块字段摊开使用。
- 2026-03-26: `auth_module.go` 已引入 `authModuleDeps`，收口 `users / tokenService / profileCommands / profileQueries / auditRecorder`。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestAuthModuleUsesTypedDeps|TestNewRouterRegistersStudentChallengeRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/auth/... -count=1`
