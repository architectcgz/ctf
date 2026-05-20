# Challenge / Ops Boundary Phase 3 Slice 2 Implementation Plan

## Objective

完成后端模块边界迁移里 `challenge -> ops` 发布通知链路的事件化收口：

- 去掉 `challenge` publish check 完成后对通知发送服务的同步调用
- 改为 `challenge` 发布自检完成事件，由 `ops` 侧复用现有通知服务作为事件消费者发送站内通知与 WebSocket 推送
- 不改变题目自检成功/失败后的用户可见通知内容、链接和发送对象

## Non-goals

- 不处理 `contest -> ops` 实时广播事件化
- 不改题目发布自检的调度、容器探活和落库逻辑
- 不重构 `ops` 通知服务的批量发送、已读、后台公告等其他能力
- 不处理 `check-workflow-complete.sh` 对 `router.go` 的 API 误报规则

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/app/composition/challenge_module.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`

## Current Baseline

- `challenge` 在 publish check 完成时通过 `ChallengeNotificationSender` 同步调用通知发送服务。
- `app/composition/challenge_module.go` 直接 new `opscmd.NotificationService`，并显式依赖 `opsinfra.NewNotificationRepository` 与 `ops.WebSocketManager`。
- `ops` 侧已经有 `RegisterPracticeEventConsumers` 模式，说明“事件消费后发通知”这条路径已存在并可复用。

## Chosen Direction

直接把题目发布自检结果改成 owner 事件：

1. 在 `challenge/contracts` 定义稳定事件 contract
2. `challenge` command service 持有 `platformevents.Bus`，在 publish check job 完成后弱发布事件
3. `challenge/runtime` 与 `app/composition` 只传入事件总线，不再传通知 sender
4. `ops.NotificationService` 新增 challenge 事件消费者，复用现有 `SendNotification` 实现实际通知
5. 删除 composition 里临时的 challenge->ops notification bridge

## Ownership Boundary

- `challenge`
  - 负责：发布“题目发布自检已完成”的业务事实事件
  - 不负责：直接拼装或调用通知基础设施实现
- `ops`
  - 负责：消费 challenge 事件并发送站内通知 / WebSocket 推送
  - 不负责：反向要求 challenge 持有通知服务依赖
- `app/composition`
  - 负责：传递共享事件总线
  - 不负责：继续为 challenge 组装临时 notification sender

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-12-challenge-ops-boundary-phase3-slice2-implementation-plan.md`
- Add: `.harness/reuse-decisions/challenge-ops-boundary-phase3-slice2.md`
- Add: `code/backend/internal/module/challenge/contracts/events.go`
- Modify: `code/backend/internal/app/composition/challenge_module.go`
- Modify: `code/backend/internal/app/router.go`
- Modify: `code/backend/internal/app/full_router_integration_test.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/module/ops/application/commands/notification_service.go`
- Modify: `code/backend/internal/module/ops/application/commands/notification_service_test.go`
- Modify: `docs/architecture/backend/{01-system-architecture.md,07-modular-monolith-refactor.md}`
- Modify: `docs/design/backend-module-boundary-target.md`

## Task Slices

### Slice 1: 让 challenge 发布事件而不是同步调通知

目标：

- 删除 `ChallengeNotificationSender`
- `challenge/runtime` 改为注入 `Events`
- publish check 完成时发布 challenge 事件

Validation:

- `cd code/backend && rg -n "ChallengeNotificationSender|SendChallengePublishCheckResult|Notifications:" internal/app/composition internal/module/challenge`
- `cd code/backend && go test ./internal/module/challenge/application/commands -run 'PublishCheck|ChallengeService.*Context' -count=1`

Review focus:

- publish check 完成是否仍保持原有通知 payload 事实
- 发布事件时是否沿用调用链上的 ctx，而不是新建 background context

### Slice 2: 让 ops 消费 challenge 事件并发送通知

目标：

- `NotificationService` 注册 challenge 事件消费者
- 复用现有 `SendNotification`，保持通知标题、内容、链接一致

Validation:

- `cd code/backend && go test ./internal/module/ops/application/commands -run 'Register.*EventConsumers|SendNotification' -count=1`

Review focus:

- `ops` 是否只消费稳定 contract，而不是反向依赖 challenge command 实现
- challenge 和 practice 的事件消费者是否都继续可用

### Slice 3: 文档与 composition 对齐

目标：

- 当前架构事实改成 `challenge` 发布事件、`ops` 消费事件
- app/composition 不再保留 challenge notification bridge

Validation:

- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`

Review focus:

- 文档是否明确这是 phase 3 第二刀，而不是“ops 事件化整体完成”

## Integration Checks

- 题目发布自检成功后仍通知申请发布的教师
- 题目发布自检失败后仍通知申请发布的教师，并保留 failure summary
- challenge 不再直接构造或调用通知服务
- ops 通知服务同时保留 practice 与 challenge 事件消费者

## Rollback / Recovery Notes

- 本切片无 schema 变更，可代码级整体回退
- 若事件通知验证失败，应整体回退这刀，而不是恢复一半 composition bridge

## Risks

- 如果事件 payload 字段选得不稳定，后续会让 ops 消费者继续承接 challenge 内部细节
- 如果只删同步 sender、不补 ops 消费者，发布自检通知会静默丢失
- `router.go` 改 builder 签名后，`check-workflow-complete.sh` 里现有的 API surface 规则可能继续误报

## Verification Plan

1. `cd code/backend && go test ./internal/module/challenge/application/commands -run 'PublishCheck|ChallengeService.*Context' -count=1`
2. `cd code/backend && go test ./internal/module/ops/application/commands -run 'Register.*EventConsumers|SendNotification' -count=1`
3. `cd code/backend && go test ./internal/app -run 'TestChallengeModuleUsesTypedPortsDeps|TestRouterBuildUsesCompositionModules' -count=1`
4. `cd code/backend && go test ./internal/module -run 'TestModuleDependencyAllowlistIsCurrent' -count=1`
5. `python3 scripts/check-docs-consistency.py`
6. `bash scripts/check-consistency.sh`
7. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- owner 明确：`challenge` 只发事实，`ops` 只处理通知副作用
- reuse point 明确：复用现有 `platformevents.Bus` 和 `ops.NotificationService`
- 这刀同时解决行为与结构：不仅保留通知能力，也删掉同步 sender 与 composition bridge
- 这刀不试图顺手把 `contest -> ops` 广播一并做掉，避免把 phase 3 的不同副作用链混成一个大提交
