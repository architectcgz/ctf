<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# 跨副本事件总线与 Outbox Relay Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking; flip each checkbox immediately after the expected result is reached.

**Goal:** 把依赖进程内 `platform/events.Bus` 的关键 side effect 收口为跨副本一致的可恢复链路：DB-backed 领域事件先写 outbox，再由 dispatcher / Redis Stream fanout 驱动通知、缓存失效与本地 WebSocket 推送。

**Architecture:** 保留 `platform/events.Bus` 作为进程内模块解耦接口，但不再把它当成跨副本消息中间件。新增通用 `platform_event_outbox` 和 codec/dispatcher owner：业务成功与 outbox enqueue 在同一事务中完成；dispatcher 将 fanout 事件写入 Redis Stream；每个 API 副本用自己的 cursor 消费 stream 并执行本地 websocket/cache side effect。

**Tech Stack:** Go, PostgreSQL outbox table, GORM transaction, Redis Streams, JSON event codec, go-redis/v9, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-distributed-event-bus-and-outbox-relay`
- Started At: `2026-06-12T00:00:00Z`
- Worktree: `后续实现时运行 scripts/start-implementation.sh 2026-06-12-distributed-event-bus-and-outbox-relay 生成`
- Branch: `task/2026-06-12-distributed-event-bus-and-outbox-relay`

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

## Inputs

- Source docs:
  - `docs/plan/impl-plan/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-07-contest-realtime-relay-externalization-implementation-plan.md`
  - `docs/architecture/backend/01-system-architecture.md`
  - `docs/architecture/backend/05-key-flows.md`
- Related architecture/contracts:
  - `code/backend/internal/platform/events/bus.go`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/module/ops/application/commands/notification_service.go`
  - `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
  - `code/backend/internal/module/challenge/application/challengecatalog/published_catalog_event.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
  - `code/backend/internal/module/ops/runtime/module.go`
  - `code/backend/internal/module/ops/infrastructure/contest_realtime_stream.go`
  - `code/backend/internal/module/contest/infrastructure/realtime_outbox_repository.go`
- Related prior work:
  - `docs/plan/impl-plan/2026-06-07-contest-realtime-relay-externalization-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-08-multi-instance-distributed-lock-hardening-implementation-plan.md`

## Task Classification

- Classification: `结构性改动 / 非琐碎任务`
- Why:
  - 横跨 `platform/events`、`practice`、`challenge`、`ops`、`app/composition`、PostgreSQL migration 和 Redis Stream。
  - 改变副作用可靠性边界：从 best-effort in-memory handler 变成事务 outbox + fanout。
  - 需要严格处理事务边界、payload version、幂等、dispatcher 多副本竞争和 root lifecycle。

## Files

- Create:
  - `code/backend/migrations/000018_create_platform_event_outbox.up.sql`
  - `code/backend/migrations/000018_create_platform_event_outbox.down.sql`
  - `code/backend/internal/platform/events/outbox.go`
  - `code/backend/internal/platform/events/outbox_codec.go`
  - `code/backend/internal/platform/events/outbox_repository.go` 或 infrastructure 等价文件
  - `code/backend/internal/platform/events/outbox_repository_test.go`
  - `code/backend/internal/platform/events/outbox_dispatcher.go`
  - `code/backend/internal/platform/events/outbox_dispatcher_test.go`
  - `code/backend/internal/platform/events/stream_fanout.go`
  - `code/backend/internal/platform/events/stream_fanout_test.go`
- Modify:
  - `code/backend/internal/platform/events/bus.go`
  - `code/backend/internal/app/composition/root.go`
  - `code/backend/internal/module/practice/application/commands/submission_service.go`
  - `code/backend/internal/module/practice/application/commands/manual_review_service.go`
  - `code/backend/internal/module/practice/application/commands/service_lifecycle.go`
  - `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
  - `code/backend/internal/module/challenge/application/challengepublishcheck/service.go`
  - `code/backend/internal/module/challenge/application/challengecatalog/published_catalog_event.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/repository.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
  - `code/backend/internal/module/ops/application/commands/notification_service.go`
  - `code/backend/internal/module/ops/runtime/module.go`
  - `code/backend/internal/app/composition/ops_module.go`
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
  - `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
  - `code/backend/internal/module/ops/application/commands/notification_service_test.go`
  - `code/backend/internal/module/ops/api/http/notification_http_integration_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `platform/events.Bus` 当前同步串行调用本进程 handlers，`Payload any` 仅适合进程内类型断言。
  - contest realtime 已有 `contest_realtime_outbox`、dispatcher、Redis Stream、per-instance cursor 和 dedupe marker 模式。
  - notification 当前 DB 写后直接调用本地 WebSocket manager，无法触达连接在其他副本上的用户。
  - progress cache invalidation 当前依赖 `practice.flag_accepted` in-memory subscribe handler。
- Reuse / extend / split / create-new decision:
  - 复用 contest realtime outbox/stream 的设计经验，但新增通用 platform outbox，不直接把所有事件塞进 contest 表。
  - 保留 in-memory Bus 用于非关键本进程解耦；首批关键 side effect 改走 outbox/fanout。
  - 通用事件必须有 codec/版本，不允许 `Payload any` 直接 JSON 化。
  - dispatcher / stream consumer 纳入 `ops` 或 root background lifecycle，不私自悬挂 goroutine。
- Owner boundary:
  - `platform/events`：通用 outbox entity、codec、repository interface、dispatcher/fanout substrate owner。
  - `practice`：flag accepted 领域事实和提交事务 owner。
  - `challenge`：publish check finished 领域事实和 job update 事务 owner。
  - `ops`：notification row、WebSocket fanout、stream consumer lifecycle owner。
  - `app/composition`：组装 outbox repository、dispatcher、fanout handler registry owner。
- Why this is the narrowest safe surface:
  - 只迁移已知跨副本 correctness blocker，不把全站所有弱事件一次性搬到新总线。
  - 保留现有 modular monolith 和 Redis Stream 基础设施，不引入新外部中间件。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
  - `dispatching-parallel-agents`
- Why this pass fits:
  - 事件链路涉及事务一致性和跨副本投递语义，必须先读清现有 Bus、contest stream 和 notification/progress side effect。
- grill-with-docs findings:
  - 当前 `platform/events.Bus` 明确是进程内 map + handler slice，不能承担跨副本一致性。
  - contest realtime 已证明项目接受“DB outbox + Redis Stream + local fanout”的设计。
  - notification created/read 当前不是 Bus 事件，而是 DB 写后直接本地 websocket 推送，是多副本下的直接缺口。
- Plan adjustments after challenge:
  - 首批 outbox enqueue 失败应让对应业务 transaction rollback，避免业务事实成功但关键 side effect 永久丢失。
  - Redis Stream fanout 只保证副本级消费，不承诺客户端离线 replay。
  - dispatcher claim 必须处理多副本重复处理风险，不能照搬 contest outbox 中未加 claim 的 `ListPending`。

## Ordered Task Slices

### Slice 1: platform outbox substrate

- [ ] **Step 1: 写 outbox migration / repository 测试**
  - Create: `code/backend/internal/platform/events/outbox_repository_test.go`
  - 覆盖 enqueue、list due、claim 不重复、mark dispatched、mark failed/backoff、UTC timestamp。

- [ ] **Step 2: 新增 platform_event_outbox migration**
  - Create: `code/backend/migrations/000018_create_platform_event_outbox.up.sql`
  - Create: `code/backend/migrations/000018_create_platform_event_outbox.down.sql`
  - 字段包含 event_name、payload、payload_version、route、dedupe_key、status、attempt_count、next_attempt_at、locked_by、locked_until、occurred_at、created_at、updated_at。

- [ ] **Step 3: 实现 outbox repository**
  - Create: `code/backend/internal/platform/events/outbox.go`
  - Create: `code/backend/internal/platform/events/outbox_repository.go`
  - `ListDue/Claim` 使用 PostgreSQL row lock 或 lease，避免多副本 dispatcher 重复处理。

- [ ] **Step 4: 写 codec 测试**
  - Create: `code/backend/internal/platform/events/outbox_codec.go`
  - 测试 encode/decode `practice.flag_accepted`、`challenge.publish_check_finished`、`notification.created`、`notification.read`，未知 event/version 失败。

- [ ] **Step 5: 运行 platform events substrate tests**
  - Run: `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec' -count=1`

### Slice 2: Redis Stream fanout substrate

- [ ] **Step 6: 写 generic stream fanout 测试**
  - Create: `code/backend/internal/platform/events/stream_fanout_test.go`
  - 覆盖 publish/consume、per-instance cursor、fanout 前失败不推进 cursor、dedupe marker。

- [ ] **Step 7: 抽通用 stream fanout adapter**
  - Create: `code/backend/internal/platform/events/stream_fanout.go`
  - 参考 `ops/infrastructure/contest_realtime_stream.go` 的 XADD、Lua dedupe marker、XREAD cursor 模式。

- [ ] **Step 8: 实现 outbox dispatcher**
  - Create: `code/backend/internal/platform/events/outbox_dispatcher.go`
  - route=stream 写 Redis Stream；失败 mark failed + backoff；成功 mark dispatched。

- [ ] **Step 9: 接入 root lifecycle**
  - Modify: `code/backend/internal/app/composition/root.go`
  - Modify: `code/backend/internal/module/ops/runtime/module.go`
  - Dispatcher / consumer 由 root background job 管理，ctx 从 root 传入。

### Slice 3: migrate practice/challenge domain events

- [ ] **Step 10: 写 practice transaction outbox 测试**
  - Modify: `code/backend/internal/module/practice/application/commands/submission_flag_side_effects_test.go`
  - 正确提交写 submission 与 `practice.flag_accepted` outbox 同事务；重复解出不写第二条；outbox enqueue 失败 rollback。

- [ ] **Step 11: 改 practice flag accepted enqueue**
  - Modify: `code/backend/internal/module/practice/application/commands/submission_service.go`
  - Modify: `code/backend/internal/module/practice/application/commands/manual_review_service.go`
  - 用 repository transaction 写业务事实 + outbox；保留 weak bus 兼容时须明确只作本进程非关键通知。

- [ ] **Step 12: 写 challenge publish check outbox 事务测试**
  - Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
  - Modify/Create: `code/backend/internal/module/challenge/infrastructure/challenge_command_repository_test.go`
  - 覆盖 job final update 与 `challenge.publish_check_finished` outbox 在同一 `*gorm.DB` transaction handle 内写入。
  - 断言 outbox enqueue 失败时，publish check job final status / published_at 更新回滚。

- [ ] **Step 13: 补 challenge publish check transaction owner**
  - Modify: `code/backend/internal/module/challenge/ports/ports.go`
  - Modify: `code/backend/internal/module/challenge/infrastructure/repository.go`
  - Modify: `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
  - 当前 `Repository.WithinTransaction(ctx, fn)` 与 `WithDB(tx)` 已存在；把它显式暴露到 publish check service 可用的 port / command repository 层。
  - outbox repository 必须支持绑定同一个 `*gorm.DB` transaction，不能在 service 内另开独立 DB handle。

- [ ] **Step 14: 改 challenge publish check enqueue**
  - Modify: `code/backend/internal/module/challenge/application/challengepublishcheck/service.go`
  - `finishPublishCheckJob` 不再先 `UpdatePublishCheckJob` 再 weak publish；改为在 transaction 中更新 job final fields 并 enqueue `challenge.publish_check_finished` outbox event。
  - Payload 增加 `OccurredAt` 或由 outbox event metadata 提供 UTC occurred_at。

### Slice 4: migrate notification and progress fanout

- [ ] **Step 15: 写 notification fanout 测试**
  - Modify: `code/backend/internal/module/ops/application/commands/notification_service_test.go`
  - `SendNotification` / `MarkAsRead` 写 DB row + fanout outbox，不再直接依赖本地 manager。

- [ ] **Step 16: 改 notification service**
  - Modify: `code/backend/internal/module/ops/application/commands/notification_service.go`
  - Create / CreateBatch / MarkAsRead 在 DB transaction 内 enqueue `notification.created/read` stream events。

- [ ] **Step 17: 写 stream consumer local fanout 测试**
  - Modify/Create: `code/backend/internal/module/ops/application/commands/notification_service_test.go`
  - 消费 `notification.created/read` 后才调用 WebSocket manager；用户连接在另一个副本也可通过 stream 收到。

- [ ] **Step 18: 改 progress invalidation handler**
  - Modify: `code/backend/internal/module/practice/application/queries/progress_timeline_service.go`
  - 从 in-memory subscribe 迁移到 stream fanout handler；删除 Redis cache 幂等。

### Slice 5: verification and docs

- [ ] **Step 19: 更新架构事实源**
  - Modify: `docs/architecture/backend/01-system-architecture.md`
  - Modify: `docs/architecture/backend/05-key-flows.md`
  - 写清 in-memory Bus、DB outbox、Redis Stream fanout 三者边界。

- [ ] **Step 20: 运行最小验证**
  - Run: `cd code/backend && go test ./internal/platform/events ./internal/module/ops/... ./internal/module/practice/... ./internal/module/challenge/... -run 'Outbox|Stream|Notification|FlagAccepted|PublishCheck|Progress' -count=1`

- [ ] **Step 21: Commit**
  - Run: `git add code/backend/migrations code/backend/internal/platform/events code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/module/ops code/backend/internal/app/composition docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md && git commit -m "feat(backend): 增加跨副本事件 outbox relay" -m "新增 platform event outbox 与 Redis Stream fanout，迁移通知、进度失效和首批领域事件的关键 side effect。" -m "明确 in-memory Bus 只保留进程内解耦语义，跨副本一致性由 outbox/stream owner 承担。" -m "Task: 2026-06-12-distributed-event-bus-and-outbox-relay"`

## Validation

- Commands:
  - `cd code/backend && go test ./internal/platform/events -count=1`
  - `cd code/backend && go test ./internal/module/ops/application/commands -run Notification -count=1`
  - `cd code/backend && go test ./internal/module/ops/infrastructure -run Stream -count=1`
  - `cd code/backend && go test ./internal/module/practice/application/commands -run 'FlagAccepted|Submission|ManualReview' -count=1`
  - `cd code/backend && go test ./internal/module/practice/application/queries -run Progress -count=1`
  - `cd code/backend && go test ./internal/module/challenge/... -run 'PublishCheck|PublishedCatalog' -count=1`
  - `git diff --check -- code/backend/migrations code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md`
- Manual checks:
  - 两个 API 副本：用户 WebSocket 连接在副本 B，副本 A 创建 notification，B 能收到 `notification.created`。
  - 副本 A 提交 flag accepted，副本 B 的 progress cache invalidation handler 能执行。
  - dispatcher 中断后重启，pending outbox 继续投递。
  - stream consumer 在 fanout 前退出后，重启仍能重读并 fanout。
- Review focus:
  - 业务事实与 outbox enqueue 是否同事务。
  - Payload codec 是否 versioned、typed、UTC，不再依赖 `Payload any`。
  - Dispatcher claim 是否能防多副本重复处理。
  - Redis Stream fanout 是否只承诺副本级恢复，没有误承诺客户端 replay。
  - `platform/events.Bus` 边界是否清楚保留为进程内解耦。
