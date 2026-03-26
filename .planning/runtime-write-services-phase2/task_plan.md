# Task Plan

## Goal

继续推进 `runtime` Phase 2：把仍滞留在 root `application` 的写侧/运维侧服务下沉到 `application/commands`，先收口 `provisioning / cleanup / maintenance` 三块。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点剩余 root write-side 服务 | completed | 已确认 `ProvisioningService / RuntimeCleanupService / RuntimeMaintenanceService` 仍由 composition、tests 与 practice flow 直接依赖 |
| 2. 下沉服务到 `runtime/application/commands` | completed | `provisioning / cleanup / maintenance` 及 typed-nil helper/tests 已迁入 `commands` |
| 3. 切换调用点 | completed | `composition`、`practice flow`、`runtime/service_test.go`、`practice` 受影响测试已切到 `runtimecmd.New*` |
| 4. 清理 root legacy service | completed | root `runtime/application` 已删除对应 legacy 服务文件，仅保留共享 alias/helper 层 |
| 5. focused 验证 | completed | 已完成 `runtime/...`、`internal/app` runtime composition 测试、`practice/...` 与 practice flow 集成回归 |

## Acceptance Checks

- `runtime/application/commands` 新增 `provisioning / cleanup / maintenance` 写侧服务
- `composition` 与受影响测试不再 `New*` root `runtime/application` 的上述服务
- root `runtime/application` 不再保留 `provisioning_service.go`、`runtime_cleanup_service.go`、`runtime_maintenance_service.go`
- focused tests 通过

## Result

- 不改外部 API 与运行时行为
- `runtime/application` root 继续收缩为共享 helper / alias 层；写侧 owner 进一步集中到 `commands`
