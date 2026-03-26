# Progress

- 2026-03-26: 启动 `runtime-bridge-phase2`，目标是收口 runtime 与 practice/contest 之间的 composition bridge。
- 2026-03-26: `practice` runtime adapter 与拓扑请求/响应映射已从 [`practice_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/practice_module.go) 下沉到 [`runtime_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/runtime_module.go)。
- 2026-03-26: `runtime` 对外暴露的 cross-module 依赖已改用 `practiceports.InstanceRepository / RuntimeInstanceService` 与 `contestports.AWDContainerFileWriter`，不再反向依赖 `contest/infrastructure`。
- 2026-03-26: 已验证
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/contest/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/practice/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies|TestPracticeModuleUsesTypedPortsDeps|TestPracticeModuleAvoidsRuntimeBridgeGlue|TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
