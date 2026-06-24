<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# 跨副本事件总线与 Outbox Relay Implementation Plan

**Goal:** 把依赖进程内 `platform/events.Bus` 的关键 side effect 收口为跨副本一致的可恢复链路：DB-backed 领域事件先写 outbox，再由 dispatcher / Redis Stream fanout 驱动通知、缓存失效与本地 WebSocket 推送。

**Architecture:** 保留 `platform/events.Bus` 作为进程内模块解耦接口，但不再把它当成跨副本消息中间件。新增通用 `platform_event_outbox` 和 typed codec/dispatcher owner：业务成功与 outbox enqueue 在同一事务中完成；dispatcher 将 fanout 事件写入 Redis Stream；每个 API 副本用自己的 cursor 消费 stream 并执行本地 websocket/cache side effect。

**Tech Stack:** Go, PostgreSQL outbox table, GORM transaction, Redis Streams, JSON event codec, go-redis/v9, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-distributed-event-bus-and-outbox-relay`
- Parent Task Group: `2026-06-12-true-ha-group` <!-- 独立任务写"无"；task group slice 写 parent group slug -->
- Slice Index: `3/5` <!-- 独立任务写"-"；task group slice 写 "1/5"、"2/5" -->
- Depends On: `2026-06-12-redis-sentinel-and-postgres-ha-connectivity` <!-- 前置依赖 task slug，多个用逗号分隔；无依赖写"无" -->
- Started At: `2026-06-12T10:13:27Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-distributed-event-bus-and-outbox-relay`
- Branch: `task/2026-06-12-distributed-event-bus-and-outbox-relay`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `review-passed` <!-- draft | ready-for-implementation | implemented | review-pending | review-passed | archived -->
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - 新增通用 outbox 表、repository、codec、dispatcher 和 Redis Stream fanout substrate。
  - 首批迁移 `practice.flag_accepted`、`challenge.publish_check_finished`、`notification.created`、`notification.read`、user progress cache invalidation。
  - 让用户连接在任意 API 副本时，都能收到其他副本写入的 notification fanout。
  - 让 progress cache invalidation 不再依赖产生 flag 事件的那个进程内 handler。
  - 明确事件 payload JSON schema / version / UTC 时间字段，不再跨进程传 `Payload any`。
- Non-Goals:
  - 不一次性迁移所有 in-memory Bus 事件；contest realtime 已有专门 outbox / stream，是否统一到 platform outbox 另起任务。
  - 不引入 Kafka、NATS 或新的独立消息服务。
  - 不实现客户端级 WebSocket cursor / replay 协议；本任务只做副本级 fanout。
  - 不把所有后台任务迁成独立 worker 服务；dispatcher 仍在 modular monolith root lifecycle 内运行。

## Problem Statement

- Current behavior / structure:
  - `platform/events.Bus` 是进程内 `map[string][]Handler`，`Payload any` 只适合同进程同步订阅，不能覆盖多副本事件传播。
  - `practice` 在 `SubmitFlag` / `ReviewManualReviewSubmission` 成功后直接 weak publish `practice.flag_accepted`；`progress` cache invalidation 和通知依赖本地订阅者。
  - `challenge publish check` 在 job final update 后 weak publish `challenge.publish_check_finished`，事务边界和事件持久化没有同一个 owner。
  - `notification` 写 DB 后直接调用本地 websocket manager，用户连接在其他 API 副本时无法收到事件。
- Target behavior / structure:
  - 关键跨副本 side effect 改为 typed outbox event，在业务事务内和领域事实一起提交。
  - 独立 dispatcher 把待发送 outbox event fanout 到 Redis Stream，每个 API 副本本地消费后执行 websocket/cache invalidation。
  - `events.Bus` 保留进程内弱事件语义；跨副本一致性 owner 收口到 outbox/stream，不再混用。
- Why this task is needed now:
  - 真正 HA 控制面在该 slice 启动时，剩余工作集中在 T4/T5；当前 T4 已合入，T5 已通过 review / governance 并等待合并。
  - 现在的 notification/progress 路径在多副本下有直接 correctness 缺口，不能继续建立在同进程广播上。

## Inputs

- Source docs:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/distributed-event-bus-and-outbox-relay.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-07-contest-realtime-relay-externalization-implementation-plan.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
- Related architecture/contracts:
  - `code/backend/internal/platform/events/bus.go`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/module/ops/application/commands/notification_service.go`
  - `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
  - `code/backend/internal/module/practice/application/commands/submission_service.go`
  - `code/backend/internal/module/practice/application/commands/manual_review_service.go`
  - `code/backend/internal/module/challenge/application/challengepublishcheck/service.go`
  - `code/backend/internal/module/challenge/application/challengecatalog/published_catalog_event.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/repository.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
  - `code/backend/internal/module/ops/runtime/module.go`
  - `code/backend/internal/module/ops/infrastructure/contest_realtime_stream.go`
  - `code/backend/internal/module/contest/infrastructure/realtime_outbox_repository.go`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-07-contest-realtime-relay-externalization-implementation-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-08-multi-instance-distributed-lock-hardening-implementation-plan.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/shared-storage-owner-convergence.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 横跨 `platform/events`、`practice`、`challenge`、`ops`、`app/composition`、PostgreSQL migration 和 Redis Stream。
  - 改变副作用可靠性边界：从 best-effort in-memory handler 变成事务 outbox + fanout。
  - 需要明确事务 owner、payload version、幂等与 dispatcher 多副本竞争策略。

## Files

- Create:
  - `code/backend/migrations/000018_create_platform_event_outbox.up.sql`
  - `code/backend/migrations/000018_create_platform_event_outbox.down.sql`
  - `code/backend/internal/platform/events/outbox.go`
  - `code/backend/internal/platform/events/outbox_codec.go`
  - `code/backend/internal/platform/events/outbox_repository.go`
  - `code/backend/internal/platform/events/outbox_repository_test.go`
  - `code/backend/internal/platform/events/outbox_dispatcher.go`
  - `code/backend/internal/platform/events/outbox_dispatcher_test.go`
  - `code/backend/internal/platform/events/stream_fanout.go`
  - `code/backend/internal/platform/events/stream_fanout_test.go`
- Modify:
  - `code/backend/internal/platform/events/bus.go`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/app/composition/ops_module.go`
  - `code/backend/internal/module/practice/application/commands/submission_service.go`
  - `code/backend/internal/module/practice/application/commands/manual_review_service.go`
  - `code/backend/internal/module/practice/application/commands/service.go`
  - `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
  - `code/backend/internal/module/challenge/application/challengepublishcheck/service.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/repository.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
  - `code/backend/internal/module/ops/application/commands/notification_service.go`
  - `code/backend/internal/module/ops/ports/notification.go`
  - `code/backend/internal/module/ops/infrastructure/notification_repository.go`
  - `code/backend/internal/module/ops/runtime/module.go`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
- Review:
  - `code/backend/internal/module/contest/infrastructure/realtime_outbox_repository.go`
  - `code/backend/internal/module/ops/infrastructure/contest_realtime_stream.go`
  - `code/backend/internal/infrastructure/websocket/manager.go`
- Test:
  - `code/backend/internal/platform/events/*_test.go`
  - `code/backend/internal/module/practice/application/commands/*_test.go`
  - `code/backend/internal/module/practice/application/queries/*_test.go`
  - `code/backend/internal/module/challenge/application/challengepublishcheck/*_test.go`
  - `code/backend/internal/module/challenge/infrastructure/*_test.go`
  - `code/backend/internal/module/ops/application/commands/notification_service_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `platform/events.Bus` 当前同步串行调用本进程 handlers，`Payload any` 仅适合进程内类型断言。
  - contest realtime 已有 outbox、dispatcher、Redis Stream、per-instance cursor 和 dedupe marker 模式。
  - notification 当前 DB 写后直接调用本地 WebSocket manager，无法触达连接在其他副本上的用户。
  - progress cache invalidation 当前依赖 `practice.flag_accepted` in-memory subscribe handler。
  - challenge repository 已有 `WithinTransaction(ctx, fn)` / `WithDB(tx)`，可作为 publish-check 事务 owner 基础。
- Reuse / extend / split / create-new decision:
  - 复用 contest realtime outbox/stream 的设计经验，但新增通用 platform outbox，不直接把所有事件塞进 contest 专用表。
  - 保留 in-memory Bus 用于非关键本进程解耦；首批关键 side effect 改走 outbox/fanout。
  - 通用事件必须有 codec/版本，不允许 `Payload any` 直接 JSON 化。
  - dispatcher / stream consumer 纳入 root background lifecycle，不私自悬挂 goroutine。
  - practice 当前没有显式提交事务 manager，本轮先对“业务事实 + outbox”收口最小事务 owner，不顺手重构无关 command surface。
- Owner boundary:
  - `platform/events`：通用 outbox entity、codec、repository interface、dispatcher/fanout substrate owner。
  - `practice`：flag accepted 领域事实和提交事务 owner。
  - `challenge`：publish check finished 领域事实和 job final update 事务 owner。
  - `ops`：notification row、WebSocket fanout、stream consumer lifecycle owner。
  - `app/composition`：组装 outbox repository、dispatcher、fanout handler registry owner。
- Why this is the narrowest safe surface:
  - 只迁移已知跨副本 correctness blocker，不把全站所有弱事件一次性搬到新总线。
  - 保留现有 modular monolith 和 Redis Stream 基础设施，不引入新外部中间件。
  - 当前触及的结构债只限于事务 owner 和 fanout boundary；没有借机做模块大拆分。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 事件链路涉及事务一致性和跨副本投递语义，必须先读清现有 Bus、contest stream 和 notification/progress side effect。
- grill-with-docs findings:
  - 当前 `platform/events.Bus` 明确是进程内 map + handler slice，不能承担跨副本一致性。
  - contest realtime 已证明项目接受“DB outbox + Redis Stream + local fanout”的设计。
  - notification created/read 当前不是 Bus 事件，而是 DB 写后直接本地 websocket 推送，是多副本下的直接缺口。
  - challenge publish check 的最终 job update 与 event publish 不是同一个事务 owner，需要显式收口。
- Plan adjustments after challenge:
  - 首批 outbox enqueue 失败应让对应业务 transaction rollback，避免业务事实成功但关键 side effect 永久丢失。
  - Redis Stream fanout 只保证副本级消费，不承诺客户端离线 replay。
  - dispatcher claim 必须处理多副本重复处理风险，不能照搬 contest outbox 中未加 claim 的 `ListPending`。
  - 如果 practice 事务 owner 无法在现有 repo 边界内最小收口，本轮需要先补 command-side tx capability，再写 outbox。

## Execution Slices

### Slice 1: platform outbox substrate

- Goal:
  - 建立可复用的 platform outbox + typed codec + dispatcher substrate，并用测试先锁定 claim/backoff/UTC 行为。
- Dependencies:
  - `2026-06-12-redis-sentinel-and-postgres-ha-connectivity`
- Files:
  - Create:
    - `code/backend/migrations/000018_create_platform_event_outbox.up.sql`
    - `code/backend/migrations/000018_create_platform_event_outbox.down.sql`
    - `code/backend/internal/platform/events/outbox.go`
    - `code/backend/internal/platform/events/outbox_codec.go`
    - `code/backend/internal/platform/events/outbox_repository.go`
    - `code/backend/internal/platform/events/outbox_repository_test.go`
    - `code/backend/internal/platform/events/outbox_dispatcher.go`
    - `code/backend/internal/platform/events/outbox_dispatcher_test.go`
  - Modify:
    - `code/backend/internal/platform/events/bus.go`
  - Review:
    - `code/backend/internal/module/contest/infrastructure/realtime_outbox_repository.go`
  - Test:
    - `code/backend/internal/platform/events/*_test.go`
- Steps:
  - [x] Step 1: 写 outbox repository / codec / dispatcher 的失败测试，覆盖 enqueue、claim lease、mark dispatched、mark failed/backoff、typed decode 和未知 version 失败。
  - [x] Step 2: 运行 `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec|Dispatcher' -count=1`，确认红灯来自缺失实现而不是测试夹具。
  - [x] Step 3: 实现 `platform_event_outbox` migration、entity 和 repository，支持 claim lease、retry backoff 与 UTC 时间归一化。
  - [x] Step 4: 实现 typed codec 和 dispatcher，route=stream 时输出统一 payload envelope。
  - [x] Step 5: 重新运行 `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec|Dispatcher' -count=1`，确认该 slice 变绿。
- Validation:
  - `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec|Dispatcher' -count=1`
- Review focus:
  - outbox repository 是否真的避免多副本重复 claim。
  - codec 是否 typed/versioned/UTC，而不是重新包装 `Payload any`。
- Done criteria:
  - platform outbox substrate 已可被业务模块调用，且 dispatcher 的失败恢复语义有测试覆盖。

### Slice 2: Redis Stream fanout substrate

- Goal:
  - 抽通用 stream fanout substrate，复用 contest realtime 的 cursor/dedupe 经验，但 owner 落在 `platform/events`。
- Dependencies:
  - Slice 1
- Files:
  - Create:
    - `code/backend/internal/platform/events/stream_fanout.go`
    - `code/backend/internal/platform/events/stream_fanout_test.go`
  - Modify:
    - `code/backend/internal/app/composition/root.go`
    - `code/backend/internal/module/ops/runtime/module.go`
  - Review:
    - `code/backend/internal/module/ops/infrastructure/contest_realtime_stream.go`
  - Test:
    - `code/backend/internal/platform/events/stream_fanout_test.go`
- Steps:
  - [x] Step 1: 写 stream fanout 失败测试，覆盖 publish/consume、cursor 初始化、fanout 前失败不推进 cursor、dedupe marker。
  - [x] Step 2: 运行 `cd code/backend && go test ./internal/platform/events -run 'StreamFanout' -count=1`，确认红灯。
  - [x] Step 3: 实现通用 stream fanout substrate，并在 root / ops runtime 预留 dispatcher、consumer 生命周期接线。
  - [x] Step 4: 重新运行 `cd code/backend && go test ./internal/platform/events -run 'StreamFanout' -count=1`，确认变绿。
- Validation:
  - `cd code/backend && go test ./internal/platform/events -run 'StreamFanout' -count=1`
- Review focus:
  - cursor 语义是否只在本地 fanout 成功后推进。
  - dedupe marker 是否避免重复 publish，同时不过度承诺 exactly-once。
- Done criteria:
  - 通用 stream fanout 可被 notification/progress consumer 直接复用。

### Slice 3: migrate practice and challenge domain events

- Goal:
  - 让 `practice.flag_accepted` 与 `challenge.publish_check_finished` 从 weak publish 改为事务内 outbox enqueue。
- Dependencies:
  - Slice 1
- Files:
  - Create:
    - 如需要：practice command tx helper / adapter 文件
  - Modify:
    - `code/backend/internal/module/practice/application/commands/service.go`
    - `code/backend/internal/module/practice/application/commands/submission_service.go`
    - `code/backend/internal/module/practice/application/commands/manual_review_service.go`
    - `code/backend/internal/module/challenge/application/challengepublishcheck/service.go`
    - `code/backend/internal/module/challenge/ports/ports.go`
    - `code/backend/internal/module/challenge/infrastructure/repository.go`
    - `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
  - Review:
    - `code/backend/internal/module/challenge/application/challengecatalog/published_catalog_event.go`
  - Test:
    - `code/backend/internal/module/practice/application/commands/*_test.go`
    - `code/backend/internal/module/challenge/infrastructure/*_test.go`
    - `code/backend/internal/module/challenge/application/challengepublishcheck/*_test.go`
- Steps:
  - [x] Step 1: 写 practice 事务 outbox 的失败测试，覆盖正确提交/人工审核通过与 outbox 同事务、重复解出不重复写、enqueue 失败回滚。
  - [x] Step 2: 写 challenge publish-check 事务 outbox 的失败测试，覆盖 job final update 与 outbox 同事务、enqueue 失败回滚。
  - [x] Step 3: 运行 `cd code/backend && go test ./internal/module/practice/application/commands ./internal/module/challenge/... -run 'FlagAccepted|ManualReview|PublishCheck|Outbox' -count=1`，确认红灯。
  - [x] Step 4: 补 practice/challenge 的最小事务 owner，并接入 platform outbox enqueue。
  - [x] Step 5: 重新运行相同测试，确认该 slice 变绿。
- Validation:
  - `cd code/backend && go test ./internal/module/practice/application/commands ./internal/module/challenge/... -run 'FlagAccepted|ManualReview|PublishCheck|Outbox' -count=1`
- Review focus:
  - 业务事实与 outbox enqueue 是否严格同事务。
  - challenge publish-check 是否复用现有 repository tx owner，而不是 service 内部另开 DB handle。
- Done criteria:
  - practice/challenge 两条关键弱事件已迁出 best-effort 路径。

### Slice 4: migrate notification and progress fanout

- Goal:
  - 让 notification 与 progress invalidation 真正经过 stream fanout，而不是本地 websocket / 本地 bus 订阅。
- Dependencies:
  - Slice 1
  - Slice 2
  - Slice 3
- Files:
  - Create:
    - 如需要：notification stream consumer / handler 文件
  - Modify:
    - `code/backend/internal/module/ops/application/commands/notification_service.go`
    - `code/backend/internal/module/ops/ports/notification.go`
    - `code/backend/internal/module/ops/infrastructure/notification_repository.go`
    - `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
    - `code/backend/internal/module/ops/runtime/module.go`
    - `code/backend/internal/app/composition/ops_module.go`
  - Review:
    - `code/backend/internal/infrastructure/websocket/manager.go`
  - Test:
    - `code/backend/internal/module/ops/application/commands/notification_service_test.go`
    - `code/backend/internal/module/practice/application/queries/*_test.go`
- Steps:
  - [x] Step 1: 写 notification fanout 和 progress invalidation 的失败测试，覆盖 `SendNotification` / `PublishAdminNotification` / `MarkAsRead` 的 outbox enqueue，以及消费后才触发 websocket / cache delete。
  - [x] Step 2: 运行 `cd code/backend && go test ./internal/module/ops/application/commands ./internal/module/practice/application/queries -run 'Notification|Progress|Fanout' -count=1`，确认红灯。
  - [x] Step 3: 扩展 notification repository 的事务能力与 service 的 outbox 路径，接线本地 consumer 与 websocket manager。
  - [x] Step 4: 把 progress invalidation 从 in-memory subscribe 迁到 outbox handler；当前 progress cache 是共享 Redis，删除一次即可，不做每副本重复删除。
  - [x] Step 5: 重新运行相同测试，确认该 slice 变绿。
- Validation:
  - `cd code/backend && go test ./internal/module/ops/application/commands ./internal/module/practice/application/queries -run 'Notification|Progress|Fanout' -count=1`
- Review focus:
  - notification 是否不再直接依赖当前副本 manager 发消息。
  - progress cache invalidation 是否已完全脱离进程内 bus 订阅。
- Done criteria:
  - 多副本通知和 progress invalidation owner 已收口到 stream fanout。

### Slice 5: docs and completion verification

- Goal:
  - 更新架构事实源，完成最小充分验证，为独立 review 准备 handoff。
- Dependencies:
  - Slice 1-4
- Files:
  - Create:
    - `docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md`（独立 review 通过后归档）
  - Modify:
    - `docs/architecture/backend/01-system-architecture.md`
    - `docs/architecture/backend/05-key-flows.md`
    - 本计划文件的 `Validation Evidence` / `Independent Review Handoff`
  - Review:
    - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`
  - Test:
    - 受影响 Go package 的验证命令
- Steps:
  - [x] Step 1: 更新架构事实源，写清 in-memory Bus、DB outbox、Redis Stream fanout 三者边界。
  - [x] Step 2: 运行最小充分验证与 `git diff --check`，记录结果到本计划的 `Validation Evidence`。
  - [x] Step 3: 进入独立 review gate，修复 material findings 后回填 handoff 与 residual risks。
- Validation:
  - `cd code/backend && go test ./internal/platform/events ./internal/module/ops/... ./internal/module/practice/... ./internal/module/challenge/... -run 'Outbox|Stream|Notification|FlagAccepted|PublishCheck|Progress' -count=1`
  - `cd /home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-distributed-event-bus-and-outbox-relay && git diff --check -- docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/archive/impl-plan/2026-06/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md code/backend/migrations code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/testutil/systemapp code/backend/internal/app`
- Review focus:
  - 架构文档是否如实反映当前实现边界，而不是保留老的“统一复用 bus”描述。
  - 验证证据是否完整可追溯。
- Done criteria:
  - 代码、文档、验证与 review handoff 都挂到当前 gate，并可进入 completion-full。

## Impact And Compatibility

- API / DTO:
  - 对外 HTTP API 不改协议；内部事件 payload 改成 typed JSON envelope，作为进程外契约新增版本字段。
- Data / migration:
  - 新增 `platform_event_outbox` 表；无历史数据回填要求。
- State / cache / queue / event:
  - 新增 Redis Stream fanout key 与 cursor key；progress invalidation 从本地 bus 订阅迁到 stream consumer。
- Runtime / config:
  - 复用现有 Redis 连接；如需 stream key 命名常量，落在现有 ops cachekeys / platform events owner。
- Frontend route / state / UX:
  - 无直接前端改动，但通知跨副本可见性变化属于用户可感知行为修复。
- Docs / contracts:
  - backend architecture / key flows 必须同步说明 outbox + stream 边界。

## Plan Review / Architecture Fit

- Target owner boundary:
  - `platform/events` 只 owner 通用 outbox/fanout substrate；业务 payload 组装仍在各自 bounded context。
- Reuse points / landing zones:
  - stream 读写、cursor 和 dedupe 复用 contest realtime 经验；notification/websocket 继续留在 ops；challenge tx owner 复用现有 repository transaction。
- Known structural debt touched:
  - practice command side 原本没有显式“业务提交事务 + outbox”边界；本轮已通过 command-side tx helper / adapter 把正确提交、人工审核通过与 `practice.flag_accepted` outbox 收到同一事务。
  - challenge publish-check 原本对 tx owner 暴露不够明确；review 修复后，发布通过时 challenge status update、publish-check job final update 与 `challenge.publish_check_finished` outbox enqueue 已由 `ChallengePublishCheckOutboxTxManager` 收到同一事务，weak catalog event 只在事务成功后发送。第一次独立复审又指出 stale full-row overwrite 风险；当前实现已改为 final transaction 内 `LockChallengeByID` 读取当前 challenge，并通过 `MarkChallengePublished` 只定向更新 `status` / `updated_at`。
- How this plan avoids behavior-only convergence:
  - 不只把通知消息换个发送通道，而是把事务 owner、事件编码和多副本消费边界一起收口。
- Hidden second-redesign risk:
  - 若把 outbox 先落到 ops 或 challenge 单模块，后续 practice/其他模块还会重复造轮子；因此本轮直接在 `platform/events` 建 substrate。
- Decision after review:
  - 按当前切片推进；不额外拉高为全站事件总线统一改造。

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/INDEX.md`（当前 worktree 缺失；本任务复审以当前 implementation plan、架构事实源和源码 diff 为准）
- Fact sources to update after implementation:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
- Plan-only notes that must not become architecture source:
  - 本计划里的 step 顺序、review focus、命令清单。
- Archive condition:
  - completion-full、独立 review 和 workflow governance 完成后，用 task archive 脚本归档到 `docs/plan/archive/impl-plan/2026-06/`。

## Validation Plan

- Per-slice commands:
  - `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec|Dispatcher|StreamFanout' -count=1`
  - `cd code/backend && go test ./internal/module/practice/application/commands ./internal/module/challenge/... -run 'FlagAccepted|ManualReview|PublishCheck|Outbox' -count=1`
  - `cd code/backend && go test ./internal/module/ops/application/commands ./internal/module/practice/application/queries -run 'Notification|Progress|Fanout' -count=1`
- Integration commands:
  - `cd code/backend && go test ./internal/platform/events ./internal/module/ops/... ./internal/module/practice/... ./internal/module/challenge/... -run 'Outbox|Stream|Notification|FlagAccepted|PublishCheck|Progress' -count=1`
  - `bash scripts/check-architecture.sh --full`
  - `python3 scripts/check-docs-consistency.py`
  - `cd /home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-distributed-event-bus-and-outbox-relay && git diff --check -- docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/archive/impl-plan/2026-06/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md code/backend/migrations code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/testutil/systemapp code/backend/internal/app`
- Manual checks:
  - 两个 API 副本：用户 WebSocket 连接在副本 B，副本 A 创建 notification，B 能收到 `notification.created`。
  - 副本 A 提交 flag accepted，副本 B 的 progress cache invalidation handler 能执行。
  - dispatcher 中断后重启，pending outbox 继续投递。
  - stream consumer 在 fanout 前退出后，重启仍能重读并 fanout。
- Commands intentionally skipped and why:
  - 本轮不做完整多副本联机演练自动化，因为依赖外部运行环境；只保留手工验收清单。

## Validation Evidence

- Command:
  - `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec|Dispatcher|StreamFanout' -count=1`
  - Result: PASS
  - Notes: 验证 platform outbox repository / codec / dispatcher / stream fanout substrate。
- Command:
  - `cd code/backend && go test ./internal/module/practice/application/commands ./internal/module/challenge/... -run 'FlagAccepted|ManualReview|PublishCheck|Outbox' -count=1`
  - Result: PASS
  - Notes: 验证 practice flag accepted 与 challenge publish-check finished 已走事务内 outbox enqueue，且 outbox enqueue 失败会阻止对应 final write。
- Command:
  - `cd code/backend && go test ./internal/module/ops/application/commands ./internal/module/practice/application/queries -run 'Notification|Progress|Fanout' -count=1`
  - Result: PASS
  - Notes: 验证 notification command 写路径只 enqueue stream outbox，不直接 websocket；notification fanout handler 消费 outbox 后本地推送；progress cache invalidation 已从 weak Bus 迁到 outbox handler。
- Command:
  - `cd code/backend && go test ./internal/platform/events ./internal/module/ops/... ./internal/module/practice/application/queries ./internal/module/practice/runtime -run 'OutboxHandlerRegistry|Notification|Progress|Fanout|Runtime|HTTP|Outbox' -count=1`
  - Result: PASS
  - Notes: 验证 platform outbox handler registry 支持同事件多 handler、ops runtime 注册 notification handlers、practice runtime 注册 progress handler、notification HTTP websocket 集成测试显式经过 outbox dispatcher + stream consumer。
- Command:
  - `cd code/backend && go test ./internal/app/... -run 'BuildRoot|OpsModule|PracticeModule|Background|HTTPServer|RouterComposition|TypedDeps' -count=1`
  - Result: PASS
  - Notes: 验证 composition root 已注册 platform outbox dispatcher / stream consumer background jobs，ops/practice 模块通过 root handler registry 接线。
- Command:
  - `bash scripts/check-architecture.sh --full`
  - Result: PASS
  - Notes: 验证后端 module/shared/app/test architecture boundaries，以及前端架构守卫未被本轮后端接线改动破坏。
- Command:
  - `cd code/backend && go test ./internal/platform/events ./internal/module/ops/... ./internal/module/practice/... ./internal/module/challenge/... -run 'Outbox|Stream|Notification|FlagAccepted|PublishCheck|Progress' -count=1`
  - Result: PASS
  - Notes: 重新验证 runtime 拆分后 platform outbox、ops notification fanout、practice progress invalidation 与 challenge publish-check outbox 相关测试仍通过。
- Command:
  - `git diff --check -- docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/archive/impl-plan/2026-06/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/app code/backend/migrations`
  - Result: PASS
  - Notes: 检查本任务代码、migration、架构文档和计划文件没有 whitespace error。

## Review Blocker Fix Notes

- Independent review:
  - Review record: `docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md`
  - Initial gate verdict: `blocked`
  - Final gate verdict: `pass`
  - Current plan status: 初审 4 个 blocker 和第一次复审新增的 stale publish-check blocker 均已在实现上下文修复，并通过独立只读复审。
- Blocker 1 fixed:
  - `platform_event_outbox.dedupe_key` 的 GORM `ON CONFLICT` 已补 `TargetWhere`，与 migration 里的 `WHERE dedupe_key <> ''` partial unique index 对齐。
  - Regression: `TestOutboxRepositoryDedupeConflictTargetsPartialIndex`
- Blocker 2 fixed:
  - `challengepublishcheck.finishPublishCheckJob` 现在把 challenge published 状态、publish-check job final 状态和 `challenge.publish_check_finished` outbox enqueue 放进同一个 transaction；weak catalog event 只在 transaction 成功后发送。
  - Regression: publish-check outbox enqueue failure 时 job 仍 running，challenge 仍 draft。
- Blocker 3 fixed:
  - notification handler route 使用 `notifications.source_event_key` + partial unique index 做源事件级幂等；重复处理同一 outbox 源事件不会重复创建通知，也不会重复 enqueue notification fanout。
  - Regression: `TestNotificationRepositoryCreateIfSourceEventAbsentIsIdempotent`、`TestNotificationServiceHandlePracticeFlagAcceptedOutboxEventIsIdempotent`
- Blocker 4 fixed:
  - `OutboxRepository.ClaimDue` 在第二步 claim update 中重新检查 `next_attempt_at <= now`，避免 stale read worker 绕过 future retry backoff。
  - Regression: `TestOutboxRepositoryClaimDueDoesNotBypassFutureRetryAfterStaleRead`
- Re-review Blocker 5 fixed:
  - `challengepublishcheck.finishPublishCheckJob` 在 final transaction 内通过 `LockChallengeByID` 锁定并读取当前 challenge，不再使用 self-check 前读取的 stale model 构造发布状态、catalog event 或 outbox payload。
  - `ChallengePublishCheckOutboxTxRepository` 新增 `MarkChallengePublished`，底层仓储只更新 `status` / `updated_at`，避免通过整行 `Save` 覆盖 self-check 期间的管理员编辑。
  - `ChallengePublishCheckService` 去掉 `ChallengePublisher` 依赖，publish-check passed path 不再调用 core `PublishChallenge`，事务 owner 更明确。
  - TDD red: `cd code/backend && go test ./internal/module/challenge/application/commands -run TestServiceDispatchPublishCheckJobsDoesNotOverwriteChallengeEditedDuringSelfCheck -count=1` 按预期失败，显示旧实现会把 title / description / points 覆盖回 self-check 前的值。
  - Regression: `TestServiceDispatchPublishCheckJobsDoesNotOverwriteChallengeEditedDuringSelfCheck`

## Post-Fix Validation Evidence

- Command:
  - `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec|Dispatcher|StreamFanout' -count=1`
  - Result: PASS
  - Notes: 覆盖 outbox partial unique conflict target、ClaimDue stale read/backoff、codec、dispatcher 和 stream fanout。
- Command:
  - `cd code/backend && go test ./internal/module/ops/application/commands ./internal/module/ops/infrastructure ./internal/module/ops/api/http -run 'Notification|Outbox|Fanout' -count=1`
  - Result: PASS
  - Notes: 覆盖 notification source-event idempotency、outbox fanout 与 HTTP websocket 集成路径。
- Command:
  - `cd code/backend && go test ./internal/module/challenge/... -run 'PublishCheck|Outbox' -count=1`
  - Result: PASS
  - Notes: 覆盖 publish-check final transaction、outbox enqueue failure rollback、self-check 期间编辑不被 stale model 覆盖，以及事件 payload 使用 final transaction 内当前题目事实。
- Command:
  - `cd code/backend && go test ./internal/platform/events ./internal/module/ops/... ./internal/module/practice/... ./internal/module/challenge/... -run 'Outbox|Stream|Notification|FlagAccepted|PublishCheck|Progress' -count=1`
  - Result: PASS
  - Notes: 复跑 T4 受影响后端包的集成验证集合。
- Command:
  - `bash scripts/check-architecture.sh --full`
  - Result: PASS
  - Notes: 复跑后端 / 前端架构 guardrail。
- Command:
  - `python3 scripts/check-docs-consistency.py`
  - Result: PASS
  - Notes: 复查文档引用与图表源一致性。
- Command:
  - `git diff --check -- docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/archive/impl-plan/2026-06/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/testutil/systemapp code/backend/internal/app code/backend/migrations`
  - Result: PASS
  - Notes: 复查修复后代码、migration、review 记录、架构文档和计划文件没有 whitespace error。

## Independent Review Result

- Review target:
  - `docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md`
- Independent reviewer:
  - `codex exec --sandbox read-only --cd /home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-distributed-event-bus-and-outbox-relay --output-last-message /tmp/t4-outbox-rereview-after-fix.md`
- Final composition-root reviewer:
  - `codex exec --sandbox read-only --cd /home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-distributed-event-bus-and-outbox-relay --output-last-message /tmp/t4-outbox-final-rereview.md`
- Gate verdict:
  - `pass`
- Material findings:
  - None.
- Reviewer validation:
  - `git diff --check -- docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/archive/impl-plan/2026-06/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/testutil/systemapp code/backend/internal/app code/backend/migrations`: PASS
- Validation evidence summary:
  - platform/events、ops、practice、challenge 与 app composition 相关 Go 测试已通过。
  - `bash scripts/check-architecture.sh --full` 已通过。
  - `python3 scripts/check-docs-consistency.py` 已通过。
  - `git diff --check` 已通过。
- Architecture / contract inputs:
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
  - `docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md`
  - parent input `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/distributed-event-bus-and-outbox-relay.md` 在当前 worktree 不存在；复审不要把该路径当成可读取输入。
- Known risks / review focus:
  - 独立复审确认五个 blocker 均已关闭：partial unique conflict target、publish-check transaction、notification source-event idempotency、ClaimDue backoff recheck、publish-check finalization stale full-row overwrite。
  - 独立复审确认事务 owner 在 touched surface 内收口，尤其是 publish-check passed path 的 challenge status / job final / outbox enqueue，以及 final transaction 内锁定当前 challenge + 定向更新发布状态的实现。
  - 独立复审未发现 notification/progress 旧弱路径残留 blocker。
  - final composition-root 复审确认 `router.go` 无 diff，practice outbox handler 顺序仍为 progress handler 先注册、ops notification registrar 后注册。
- Project-local checks to consider:
  - `bash scripts/check-startup-gate.sh`
  - `bash scripts/check-workflow-governance.sh`（handoff 前）

## Rollback / Recovery

- Safe revert boundary:
  - 回滚当前分支提交并执行 `000018_create_platform_event_outbox.down.sql`。
- Data / config / runtime recovery notes:
  - 回退后 Redis Stream 上残留的 platform event 消息可以忽略；新代码下线后不会继续消费。
- Irreversible operations:
  - 无不可逆数据迁移；outbox 表可删除。

## Residual Risks

- Risk:
  - 真实多副本联机行为仍需依赖环境演练确认。
- Why acceptable:
  - 本轮先用单元/集成测试锁住事务与 stream 语义，运行环境演练属于 task-group 集成验证阶段。
- Follow-up owner, if any:
  - `2026-06-12-true-ha-group` 集成验证 owner
