# Task Plan

## Goal

继续收口 `runtime` Phase 2：把仍滞留在 root `application` 的共享契约类型下沉到 `runtime/ports`，让 `composition`、`infrastructure`、`testutil` 与受影响测试不再因为类型定义依赖 `runtime/application`。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 root `application` 剩余跨包类型 | completed | 已确认主要集中在 `TopologyCreate*` 与 `ManagedContainer*` 这两组共享契约 |
| 2. 下沉共享契约到 `runtime/ports` | completed | 已新增 `ports/topology.go` 与 `ports/metrics.go`，保留应用层实现不变 |
| 3. 切换调用点 | completed | `composition`、`infrastructure`、`testutil`、focused tests 已切到依赖 `runtime/ports` |
| 4. 清理 root legacy contract 定义 | completed | 已删除 `topology_contracts.go`，并将 `contracts.go` 收紧为 alias 层 |
| 5. focused 验证 | completed | `runtime/...`、`internal/app` 与受影响模块定向测试已通过 |

## Acceptance Checks

- `runtime/ports` 新增 topology / managed-container 相关契约定义
- `runtime/infrastructure` 不再因类型定义导入 root `runtime/application`
- `runtime/testutil` 与 `composition` 的 topology 适配不再依赖 root `runtime/application` 类型
- root `runtime/application` 不再保留 `TopologyCreate*` 与 `ManagedContainer*` 定义
- focused tests 通过

## Result

- 不改外部 API 与运行时行为
- `runtime/application` 继续保留具体应用服务，但共享契约统一下沉到 `runtime/ports`
