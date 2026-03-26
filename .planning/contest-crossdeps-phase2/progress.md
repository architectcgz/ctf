# Progress

- 2026-03-26: 启动 `contest-crossdeps-phase2`，目标是收紧 `contest` composition 的跨模块依赖。
- 2026-03-26: 已补结构守卫，要求 `contestModuleDeps` 使用 `challenge/runtime` 暴露的 typed contract，而不是直接保存模块引用。
- 2026-03-26: `contest_module.go` 已切换为 `challengeCatalog / flagValidator / containerFiles` 三类 typed deps。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestBuildContestModuleDelegatesToSubBuilders|TestContestModuleDepsAvoidConcreteContestRepositories|TestContestModuleUsesTypedCrossModuleDeps|TestNewRouterRegistersStudentChallengeRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
