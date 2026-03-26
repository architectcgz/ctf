# Progress

- 2026-03-26: 启动 `identity-readmodel-composition-phase2`，目标是标准化剩余轻量 composition 模块的装配依赖。
- 2026-03-26: `identity`、`practice_readmodel`、`teaching_readmodel` 已引入 typed deps，`Build*Module` 不再 inline new concrete repository。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestIdentityModuleUsesTypedDeps|TestPracticeReadmodelModuleUsesTypedDeps|TestTeachingReadmodelModuleUsesTypedDeps|TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouterRegistersStudentChallengeRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/identity/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/practice_readmodel/... ./internal/module/teaching_readmodel/... -count=1`
