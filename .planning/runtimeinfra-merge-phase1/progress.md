# Progress

## 2026-03-23

- 创建 `runtimeinfra` 并回 `runtime` 的独立迁移计划
- 已确认当前主要改动面集中在 composition、runtime 和 architecture rules
- 已将 `engine / cleaner / acl / runtime_metrics` 物理迁移到 `runtime/infrastructure`
- 已删除 `composition/runtimeinfra_module.go`，`BuildRuntimeModule` 改为自行装配运行时依赖
- 已移除 composition 内多余的 `runtimeEngineAdapter`，由 `runtime/infrastructure.Engine` 直接满足 runtime 应用层所需接口
- 已通过定向验证：
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/module/runtime/... -count=1`
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/app -run 'TestArchitectureRulesRejectConcreteCrossModuleImports|TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouter|TestRouterBuildUsesCompositionModules' -count=1`
