# Task Plan

## Goal

将 `internal/module/runtimeinfra` 并回 `runtime`，删除组合根里单独的 `RuntimeInfraModule` 过渡层。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 `runtimeinfra` 当前职责 | completed | `engine / cleaner / acl / metrics` 已确认全部属于 `runtime` 内部 |
| 2. 下沉到 `runtime` 内部层次 | completed | 已并入 `runtime/infrastructure` |
| 3. 移除 `BuildRuntimeInfraModule` | completed | `BuildRuntimeModule` 已自行完成装配 |
| 4. 收紧架构规则 | completed | 已移除对 `runtimeinfra` 根包的放行 |
| 5. 定向验证与文档同步 | completed | focused tests 已通过 |

## Key Files

- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/app/composition/runtime_module.go`
- `code/backend/internal/app/architecture_rules_test.go`
- `code/backend/internal/module/runtime/*`

## Acceptance Checks

- 生产代码不再 import `internal/module/runtimeinfra`
- `RuntimeInfraModule` 删除
- `runtime` 对外仍只暴露现有 handler / narrow deps
- 定向测试通过：
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/module/runtime/... -count=1`
  - `GOMAXPROCS=2 go test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouter' -count=1`

## Constraints

- 不改外部 API
- 不重新引入宽接口 façade
- 不把 Docker/engine 细节重新泄漏到 `practice / contest / system`
