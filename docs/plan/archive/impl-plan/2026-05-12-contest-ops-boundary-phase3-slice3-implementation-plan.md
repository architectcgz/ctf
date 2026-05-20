# Contest / Ops Boundary Phase 3 Slice 3 Implementation Plan

## Objective

完成后端模块边界迁移里 `contest -> ops` 实时广播链路的事件化收口：

- 去掉 `contest` command service 对 `RealtimeBroadcaster` / `SendToChannel` / `SendToUser` 语义的直接依赖
- 改为 `contest` 发布稳定的 realtime 业务事件，由 `ops` 侧消费事件并复用 `WebSocketManager` 做频道/用户推送
- 不改变现有公告、榜单刷新、AWD 预览进度的 WebSocket 路径、消息类型、频道命名和 payload 结构

## Non-goals

- 不改 `/ws/contests/:id/*` 握手鉴权、ticket 消费或频道订阅入口
- 不改公告、榜单、AWD 预览本身的业务规则、落库和计算逻辑
- 不重构 `ops` 通知服务、批量通知或通知 WebSocket handler
- 不删除 `contest/ports/realtime.go` 文件本身；只收口业务服务对实时发送接口的直接依赖

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/app/router.go`
- `code/backend/internal/module/contest/runtime/module.go`
- `code/backend/internal/module/contest/application/commands/*.go`
- `code/backend/internal/module/contest/contracts/events.go`
- `code/backend/internal/module/ops/runtime/module.go`

## Current Baseline

- `contest` runtime 在 `Build()` 后通过 `BindRealtimeBroadcaster(opsModule.WebSocketManager)` 把 WebSocket 发送能力回填到多个 command service。
- `ParticipationService`、`ScoreboardAdminService`、`SubmissionService`、`AWDService` 直接持有 `contestports.RealtimeBroadcaster`，并同步发送频道或用户消息。
- `ops` 已经有基于 `platformevents.Bus` 的通知消费者模式，说明“业务模块发事实事件，ops 做副作用适配”这条路径已在仓库内成立。

## Chosen Direction

直接把 contest realtime 输出改成 owner 事件：

1. 在 `contest/contracts` 补齐公告创建/删除、榜单刷新、AWD 预览进度的稳定事件 contract
2. `contest` command service 统一持有 `platformevents.Bus`，在完成业务写入或状态变化后弱发布 realtime 事件
3. `ops` 新增 contest realtime relay service，注册事件消费者，并把事件转成既有 WebSocket envelope 发往频道或用户
4. `contest/runtime` 不再暴露 `BindRealtimeBroadcaster`，`router.go` 不再把 `ops.WebSocketManager` 反向注入 contest command service
5. `/ws/contests/:id/announcements`、`/ws/contests/:id/scoreboard`、`/ws/contests/:id/awd-preview` 继续保留现有外部契约和频道命名

## Ownership Boundary

- `contest`
  - 负责：发布“公告已创建/删除”“榜单已刷新”“AWD 预览进度已变化”的业务事实事件
  - 不负责：直接知道 WebSocket manager、频道推送实现或用户推送实现
- `ops`
  - 负责：消费 contest 事件并做频道/用户级 WebSocket 推送
  - 不负责：反向要求 contest 持有实时广播适配器
- `app/router`
  - 负责：保留 contest realtime WS 路由入口与握手 handler
  - 不负责：继续把 `ops.WebSocketManager` 注入 contest business service

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-contest-ops-boundary-phase3-slice3-implementation-plan.md`
- Add: `.harness/reuse-decisions/contest-ops-boundary-phase3-slice3.md`
- Modify: `code/backend/internal/app/router.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `code/backend/internal/module/contest/contracts/events.go`
- Modify: `code/backend/internal/module/contest/api/http/realtime_handler.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/participation_announcement_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/scoreboard_admin_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/scoreboard_admin_freeze_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/submission_scoreboard_sync.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_preview_realtime.go`
- Modify: `code/backend/internal/module/contest/application/commands/context_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/realtime_events_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Add: `code/backend/internal/module/ops/application/commands/contest_realtime_service.go`
- Add: `code/backend/internal/module/ops/application/commands/contest_realtime_service_test.go`
- Modify: `code/backend/internal/module/ops/runtime/module.go`
- Modify: `docs/architecture/backend/01-system-architecture.md`
- Modify: `docs/architecture/backend/04-api-design.md`
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task Slices

### Slice 1: contest 侧把 realtime 输出改成事件

目标：

- 给公告、榜单、AWD 预览进度补齐稳定事件 contract
- contest command service 改为弱发布事件，不再直接发送 websocket
- runtime 层改为注入 `Events`

Validation:

- `cd code/backend && go test ./internal/module/contest/application/commands -run 'Realtime|Preview|Context' -count=1 -timeout 5m`

Review focus:

- 事件 payload 是否足够表达现有 WS envelope，而没有把 `ops` 实现细节反带回 contest
- 发布事件时是否沿用现有 ctx，而不是新建 background context

### Slice 2: ops 侧消费 contest realtime 事件

目标：

- 新增 contest realtime relay service
- 复用 `WebSocketManager` 发送公告频道、榜单频道和 AWD 预览用户消息
- 保持 WS 类型、频道命名和 payload 结构不变

Validation:

- `cd code/backend && go test ./internal/module/ops/application/commands -run 'ContestRealtime|Register.*Consumers' -count=1 -timeout 5m`

Review focus:

- `ops` 是否只依赖 `contest/contracts`，而不是 contest command / runtime 实现
- channel / user 推送的 envelope 时间戳和 payload 字段是否与原逻辑一致

### Slice 3: 移除反向注入并对齐文档

目标：

- 删除 `router.go` 里的 `BindRealtimeBroadcaster` 注入
- contest runtime 不再暴露 realtime broadcaster 绑定点
- 架构事实更新为 contest 发事件、ops 做 WS adapter

Validation:

- `cd code/backend && go test ./internal/app -run 'TestBuildContestModuleDelegatesToRuntime|TestRouterBuildUsesCompositionModules' -count=1 -timeout 5m`
- `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- contest 是否已经不再持有 `RealtimeBroadcaster` 这类 adapter 语义
- 文档是否把 phase 3 的剩余 realtime debt 标记为已收口，而不是保留过时描述

## Integration Checks

- 创建公告后，订阅 `/ws/contests/:id/announcements` 的客户端仍收到 `contest.announcement.created`
- 删除公告后，订阅 `/ws/contests/:id/announcements` 的客户端仍收到 `contest.announcement.deleted`
- 正确提交或封榜/解封后，订阅 `/ws/contests/:id/scoreboard` 的客户端仍收到 `scoreboard.updated`
- AWD 预览试跑过程中，请求发起人仍能在 `/ws/contests/:id/awd-preview` 收到 `awd.preview.progress`
- contest 不再直接持有或设置 realtime broadcaster

## Rollback / Recovery Notes

- 本切片无 schema 变更，可代码级整体回退
- 如果 ops relay 消费链未验证通过，应整体回退这刀，而不是保留“一半事件化、一半直推”的混合状态

## Risks

- AWD 预览进度事件如果少带用户或 phase 字段，会导致 user-scoped WS 推送静默丢失
- 如果只改 contest 发布、不补 ops relay，实时消息会全部中断
- 如果保留 `BindRealtimeBroadcaster` 这条回填链，结构债会留在同一 touched surface 上

## Verification Plan

1. `cd code/backend && go test ./internal/module/contest/application/commands -run 'Realtime|Preview|Context' -count=1 -timeout 5m`
2. `cd code/backend && go test ./internal/module/ops/application/commands -run 'ContestRealtime|Register.*Consumers' -count=1 -timeout 5m`
3. `cd code/backend && go test ./internal/app -run 'TestBuildContestModuleDelegatesToRuntime|TestRouterBuildUsesCompositionModules' -count=1 -timeout 5m`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1 -timeout 5m`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `timeout 600 bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：`contest` 只发事实事件，`ops` 负责 WebSocket adapter
- reuse point 明确：复用已有 `platformevents.Bus`、contest 现有事件 contract 模式、ops 侧事件消费者模式、既有 `WebSocketManager`
- 这刀同时解决行为与结构：保留原有实时推送能力，同时删掉反向 broadcaster 注入
- `contest -> ops` realtime debt 是当前 phase 3 明确剩余项，这刀直接收口整条链，不把同一 port 留成 follow-up
