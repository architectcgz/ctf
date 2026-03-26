# Progress

- 2026-03-26: 启动 `practice-crossdeps-phase2`，目标是标准化 `practice` 的跨模块 composition 装配。
- 2026-03-26: 已补结构守卫，要求 `practice_module.go` 区分 ports deps、cross-module deps、handler builder。
- 2026-03-26: `practice_module.go` 已引入 `practiceModuleExternalDeps` 与 `buildPracticeHandler`，跨模块依赖不再与持久化依赖混写。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestBuildPracticeModuleDelegatesToSubBuilders|TestPracticeModuleUsesTypedPortsDeps|TestPracticeModuleUsesTypedCrossModuleDeps|TestPracticeModuleAvoidsRuntimeBridgeGlue|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/practice/... -count=1`
