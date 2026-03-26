# Progress

- 2026-03-26: 启动 `challenge-layering-phase2`，目标是延续 `contest phase2` 的做法，先拆宽 ports，再收口 composition 依赖类型并补防回退测试。
- 2026-03-26: 已删除 `challenge/ports` 中 legacy 宽 `ChallengeRepository`，改为按 challenge command/query/flag/image-usage/writeup/topology 切分窄接口。
- 2026-03-26: `challenge` application 构造依赖与 [`challenge_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/challenge_module.go) 已切换到 typed deps；补充 architecture/router 守卫测试防止回退。
- 2026-03-26: `BuildChallengeModule` 已继续拆分为本地子 builder，分别负责 image/core/flag/topology/writeup 装配，降低 composition 文件单点膨胀。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/challenge/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies|TestChallengeModuleUsesTypedPortsDeps|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies|TestChallengeModuleUsesTypedPortsDeps|TestBuildChallengeModuleDelegatesToSubBuilders|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
