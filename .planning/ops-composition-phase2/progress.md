# Progress

- 2026-03-26: 启动 `ops-composition-phase2`，目标是标准化 `ops` 的 composition 装配模式。
- 2026-03-26: 已补结构守卫，要求 `ops_module.go` 使用 typed deps 与局部 builder，不再 inline concrete repo/runtime application 依赖。
- 2026-03-26: `ops_module.go` 已引入 `opsModuleDeps / opsNotificationDeps`，`runtime` 到 `ops` 的 stats provider 适配已回收到 `runtime_module.go`。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies|TestBuildOpsModuleDelegatesToSubBuilders|TestOpsModuleUsesTypedDeps|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/ops/... -count=1`
