# Progress

## 2026-03-26

- 启动 `runtime-layering-phase2`，目标是接通 `runtime` 已有的 `application/commands` 与 `application/queries`，把 composition、HTTP handler 与关键测试依赖从旧 root `application` facade 切走。
- 初步盘点确认：
  - `runtime/application/commands` 已有实例写侧服务
  - `runtime/application/queries` 已有实例读侧、count-running 与 proxy-ticket 服务
  - 当前主装配、HTTP handler、testutil 与多处 focused tests 仍主要依赖 root `runtime/application`
- 完成 `runtime` Phase 2 首刀分层收口：
  - `runtime_module.go` 已改为装配 `runtime/application/commands` 与 `runtime/application/queries`
  - `runtime/api/http` 与 `internal/testutil/runtimeadapters` 已切到显式组合 command/query service，不再依赖 root `application` 的实例 facade
  - root `runtime/application` 已删除 `instance_service.go`、`query_service.go`、`proxy_ticket_service.go`
  - `runtime/architecture_test.go` 与 `internal/app/router_test.go` 已补防回退守卫
- 本轮 focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/runtime/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestBuildRuntimeModuleDelegatesToSubBuilders|TestRuntimeModuleUsesTypedDeps|TestRuntimeModuleUsesCommandsQueriesServices|TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestCompositionBuildersAvoidPrivateCrossModuleFields|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/ops/... ./internal/module/practice/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
