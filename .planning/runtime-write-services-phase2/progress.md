# Progress

## 2026-03-26

- 启动 `runtime-write-services-phase2`，目标是把 root `runtime/application` 的 `provisioning / cleanup / maintenance` 写侧服务下沉到 `application/commands`。
- 初步盘点确认：
  - `composition/runtime_module.go` 仍直接 `NewRuntimeCleanupService / NewRuntimeMaintenanceService / NewProvisioningService`
  - `practice_flow`、`practice` 相关测试与 `runtime/service_test.go` 也仍直接依赖 root 写侧服务
  - 这三块都属于运行时资源创建、销毁和后台维护，更适合作为 `commands` owner
- 完成写侧服务下沉：
  - `runtime/application/commands` 新增 `provisioning_service.go`、`runtime_cleanup_service.go`、`runtime_maintenance_service.go`
  - typed-nil dependency helper 与对应测试一起迁入 `commands`
  - root `runtime/application` 删除上述 legacy write-side 服务文件
- 完成调用点切换：
  - `composition/runtime_module.go` 已改为通过 `runtimecmd.NewRuntimeCleanupService / NewRuntimeMaintenanceService / NewProvisioningService` 装配
  - runtime practice adapter 的本地持有类型已切到 `runtimecmd.*`
  - `practice_flow`、`practice` 相关测试与 `runtime/service_test.go` 已切换到 `runtimecmd` 构造器
  - `router_test.go` 新增防回退守卫，禁止 composition 重新使用 root `runtime/application` 的上述构造器
- Focused 验证通过：
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/runtime/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestBuildRuntimeModuleDelegatesToSubBuilders|TestRuntimeModuleUsesTypedDeps|TestRuntimeModuleUsesCommandsQueriesServices|TestRuntimeModuleUsesExternalPortsForCrossModuleDeps|TestCompositionBuildersAvoidPrivateCrossModuleFields|TestNewRouterUsesRuntimeHandlersForInstanceRoutes' -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/practice/... -count=1`
  - `env GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
