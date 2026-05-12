# Reuse Decision

## Change type
service / runtime / event / websocket / composition

## Existing code searched
- `code/backend/internal/module/contest/application/commands/`
- `code/backend/internal/module/contest/runtime/`
- `code/backend/internal/module/contest/contracts/`
- `code/backend/internal/module/ops/application/commands/`
- `code/backend/internal/module/ops/runtime/`
- `code/backend/internal/app/`

## Similar implementations found
- `code/backend/internal/module/contest/application/commands/awd_attack_log_commands.go`
- `code/backend/internal/module/contest/application/commands/context_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- `code/backend/pkg/websocket/manager.go`

## Decision
reuse_existing

## Reason
这次不新增新的 WebSocket 基础设施，也不再保留 contest 专用 broadcaster adapter。仓库里已经有三块成熟模式可复用：`platformevents.Bus` 作为进程内弱事件总线、`contest` 里 AWD 攻击已发布稳定事件 contract、`ops.NotificationService` 已经用“注册事件消费者 -> 做副作用适配”的方式承接通知发送。最小正确方案是让 contest 复用相同的事件发布模式，把公告 / 榜单 / AWD 预览进度表达为事实事件，再由 ops 侧复用现有 `WebSocketManager` 做 relay，而不是继续把发送接口回填进 contest command service。

## Files to modify
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_test.go`
- `code/backend/internal/module/architecture_allowlist_test.go`
- `code/backend/internal/module/contest/contracts/events.go`
- `code/backend/internal/module/contest/api/http/realtime_handler.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/application/commands/participation_service.go`
- `code/backend/internal/module/contest/application/commands/participation_announcement_commands.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_service.go`
- `code/backend/internal/module/contest/application/commands/scoreboard_admin_freeze_commands.go`
- `code/backend/internal/module/contest/application/commands/submission_service.go`
- `code/backend/internal/module/contest/application/commands/submission_scoreboard_sync.go`
- `code/backend/internal/module/contest/application/commands/awd_service.go`
- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_preview_realtime.go`
- `code/backend/internal/module/contest/application/commands/context_test.go`
- `code/backend/internal/module/contest/application/commands/realtime_events_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/ops/application/commands/contest_realtime_service.go`
- `code/backend/internal/module/ops/application/commands/contest_realtime_service_test.go`
- `code/backend/internal/module/ops/runtime/module.go`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `docs/design/backend-module-boundary-target.md`
- `docs/plan/impl-plan/2026-05-12-contest-ops-boundary-phase3-slice3-implementation-plan.md`

## After implementation
- 如果这次模式稳定，可以把“contest realtime 副作用走 owner 事件，ops 做 WS relay adapter”的复用结论补进长期 reuse 历史或更上层 skill。
