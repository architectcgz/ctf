# Progress

## 2026-03-26

- 启动 `runtime-shared-contracts-phase2`，目标是把仍挂在 root `runtime/application` 的共享契约类型下沉到 `runtime/ports`。
- 初步盘点确认：
  - `TopologyCreateNode / Network / Request / Result` 仍被 `composition`、`testutil` 与 runtime tests 直接引用
  - `ManagedContainer / ManagedContainerStat` 仍被 `infrastructure` 与 runtime tests 直接引用
  - 这些类型更接近模块共享契约，而不是具体应用服务实现细节
- 完成共享契约下沉：
  - `runtime/ports` 已新增 `topology.go` 与 `metrics.go`
  - `runtime/infrastructure`、`composition`、`internal/testutil/runtimeadapters` 与 runtime tests 已切到 `runtime/ports`
  - root `runtime/application` 已删除 `topology_contracts.go`，`contracts.go` 已收紧为 alias 层
  - `runtime/architecture_test.go` 已新增对 `infrastructure -> runtime/application` 反向依赖的防回退守卫
- 本轮 focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/runtime/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestBuildRuntimeModuleDelegatesToSubBuilders|TestRuntimeModuleUsesTypedDeps|TestRuntimeModuleUsesCommandsQueriesServices|TestRuntimeModuleUsesExternalPortsForCrossModuleDeps' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/practice/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
