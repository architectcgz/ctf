# Progress

- 2026-03-26: 启动 `runtime-contract-export-phase2`，目标是消除对 `RuntimeModule` 私有嵌套字段的跨模块读取。
- 2026-03-26: 已补结构守卫，要求 `RuntimeModule` 公开暴露跨模块 contract，并禁止 `challenge / ops / practice / contest` 回退到私有字段路径。
- 2026-03-26: `runtime_module.go` 已开始公开 practice/challenge/ops/contest contract，相关 composition 消费方已切换到公开字段。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestRuntimeModuleUsesTypedDeps|TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestChallengeModuleUsesTypedPortsDeps|TestOpsModuleUsesTypedDeps|TestPracticeModuleUsesTypedCrossModuleDeps|TestContestModuleUsesTypedCrossModuleDeps|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/challenge/... ./internal/module/ops/... ./internal/module/practice/... ./internal/module/contest/... ./internal/module/runtime/... -count=1`
