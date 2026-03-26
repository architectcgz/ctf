# Progress

- 2026-03-26: 启动 `assessment-crossdeps-phase2`，目标是标准化 `assessment` 的跨模块 composition 装配。
- 2026-03-26: 已补结构守卫，要求 `assessment_module.go` 区分本模块 deps 与 `challenge` external deps。
- 2026-03-26: `assessment_module.go` 已引入 `assessmentModuleExternalDeps`，推荐服务对 challenge 的输入已从主 deps builder 中拆出。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestAssessmentModuleUsesTypedPortsDeps|TestAssessmentModuleUsesTypedCrossModuleDeps|TestBuildAssessmentModuleDelegatesToSubBuilders|TestNewRouterRegistersStudentChallengeRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/assessment/... -count=1`
