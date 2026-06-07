<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# contest realtime relay externalization Implementation Plan

**Goal:** 为 contest realtime relay 引入“可恢复事件链路”：对可事务落库的 realtime 事件使用 PostgreSQL outbox 保证“业务成功后事件最终可投递”，再用 Redis Streams + per-instance cursor 保证“事件入队后各实例最终可补读并本地 fanout”，从而解决多实例和实例重启场景下的事件丢失问题。

**Architecture:** `contest` 继续拥有领域事实和事件类型；对 DB-backed realtime 事件，contest owner 在同一事务内写入 realtime outbox，避免“业务提交了但事件没记账”。`ops` 继续拥有 relay adapter owner：后台 dispatcher 把 outbox 记录推进 Redis Streams，各实例用自己的 cursor 从同一条 stream 补读并在 fanout 成功后推进 offset，保证每个实例都能收到所有 relay 且重启后可继续追赶。第一刀先做实例级恢复，不把公告/榜单的客户端重连补发协议一并拉进来。

**Tech Stack:** Go, PostgreSQL outbox table, GORM transaction, go-redis/v9 Streams, Redis cursor state, WebSocket manager, focused Go tests

---

## Task Metadata

- Task Slug: `2026-06-07-contest-realtime-relay-externalization`
- Started At: `2026-06-07T09:22:34Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-07-contest-realtime-relay-externalization`
- Branch: `task/2026-06-07-contest-realtime-relay-externalization`

## Objective And Non-Goals

- Objective:
  - 保留 `code/backend/internal/platform/events/bus.go` 作为进程内模块解耦总线，不把它升级成跨进程消息中间件。
  - 为 contest realtime 引入可持久化的 outbox 记录，至少覆盖可以和业务写事务一起提交的 DB-backed realtime 事件。
  - 在 `ops` runtime 内新增 lifecycle-managed dispatcher，把 outbox 记录推进 Redis Streams。
  - 用 Redis Streams + per-instance cursor + 重试退避收口“事件入队后实例没补读/本地没 fanout 成功”的丢失窗口。
  - 为无法成功消费的 relay event 定义死信/告警落点，不再静默吞掉。
- Non-Goals:
  - 不改 notification WebSocket 链路，不把通知也一起外置。
  - 第一刀不改前端 `useWebSocket()` 的 cursor / resume 协议，也不为公告/榜单新增客户端级补发契约。
  - 不改 `assessment`、`practice` 等其他事件 consumer 的 owner 边界。
  - 不引入 Kafka、NATS 或新的独立消息基础设施。
  - 不把 scoreboard 通道改成整榜数据推送，仍保持“事件触发 + HTTP 刷新”模式。
  - 不承诺第一刀内完成 `AWD 预览` 的客户端断线补发；这件事单独视为后续 capability。

## Inputs

- Source docs:
  - `docs/plan/README.md`
  - `docs/文档规范.md`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/frontend/05-websocket-composables.md`
  - `docs/design/backend-module-boundary-target.md`
  - `code/backend/internal/platform/events/bus.go`
  - `code/backend/internal/module/contest/contracts/events.go`
  - `code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go`
  - `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
  - `code/backend/internal/module/ops/runtime/module.go`
  - `code/backend/internal/module/ops/application/commands/contest_realtime_service.go`
  - `code/backend/internal/infrastructure/websocket/manager.go`
  - `code/backend/internal/module/contest/api/http/realtime_handler.go`
- Related prior work:
  - `code/backend/internal/module/ops/application/commands/contest_realtime_service_test.go`
  - `code/backend/internal/module/contest/application/commands/realtime_events_test.go`
  - `code/backend/internal/module/contest/application/commands/context_test.go`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 横跨 `contest -> ops -> app/composition` 多个模块边界。
  - 同时触达 PostgreSQL 事务、Redis Streams、后台任务生命周期、WebSocket relay 和事件 consumer owner。
  - 命中活动技术债：`assessment / ops` 的副作用事件化边界仍需继续收口。

## Files

- Create:
  - `docs/plan/impl-plan/2026-06-07-contest-realtime-relay-externalization-implementation-plan.md`
  - `code/backend/migrations/000013_create_contest_realtime_outbox.up.sql`
  - `code/backend/migrations/000013_create_contest_realtime_outbox.down.sql`
  - `code/backend/internal/module/contest/entity/realtime_outbox.go`
  - `code/backend/internal/module/contest/ports/realtime_outbox.go`
  - `code/backend/internal/module/contest/infrastructure/realtime_outbox_repository.go`
  - `code/backend/internal/module/contest/infrastructure/realtime_outbox_repository_test.go`
  - `code/backend/internal/module/ops/ports/realtime.go`
  - `code/backend/internal/module/ops/infrastructure/cachekeys/redis_keys.go`
  - `code/backend/internal/module/ops/infrastructure/contest_realtime_stream.go`
  - `code/backend/internal/module/ops/infrastructure/contest_realtime_stream_test.go`
- Modify:
  - `code/backend/internal/module/ops/application/commands/contest_realtime_service.go`
  - `code/backend/internal/module/ops/application/commands/contest_realtime_service_test.go`
  - `code/backend/internal/module/ops/runtime/module.go`
  - `code/backend/internal/app/composition/ops_module.go`
  - `code/backend/internal/module/contest/application/commands/realtime_broadcast.go`
  - `code/backend/internal/module/contest/application/commands/context_test.go`
  - `code/backend/internal/module/contest/application/commands/participation_announcement_commands.go`
  - `code/backend/internal/module/contest/application/commands/submission_scoreboard_sync.go`
  - `code/backend/internal/module/contest/application/commands/scoreboard_admin_freeze_commands.go`
  - `code/backend/internal/module/contest/application/commands/awd_preview_realtime.go`
  - `code/backend/internal/module/contest/application/commands/realtime_events_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`
- Review:
  - `docs/reviews/backend/2026-06-07-contest-realtime-relay-externalization-review.md`
- Test:
  - `code/backend/internal/module/ops/application/commands/contest_realtime_service_test.go`
  - `code/backend/internal/module/ops/infrastructure/contest_realtime_stream_test.go`
  - `code/backend/internal/module/contest/infrastructure/realtime_outbox_repository_test.go`
  - `code/backend/internal/module/contest/application/commands/realtime_events_test.go`
  - `code/backend/internal/module/contest/application/commands/context_test.go`
  - `code/backend/internal/app/router_composition_structure_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - `code/backend/internal/module/ops/runtime/module.go`
  - `code/backend/internal/app/composition/contest_module.go`
  - `code/backend/internal/module/assessment/runtime/module.go`
  - `code/backend/internal/module/practice/runtime/module.go`
  - `code/backend/internal/module/contest/api/http/realtime_handler.go`
  - `code/backend/internal/infrastructure/websocket/manager.go`
- Reuse / extend / split / create-new decision:
  - 复用现有 `platformevents.Bus`、contest 事件契约、WebSocket channel 命名和 `WebSocketManager`；复用 `contest status transition` 已有的“事务内记账 + 后台 replay side effects”思路；新增 contest realtime outbox 和 ops stream relay，而不是把分布式语义塞回 `contest` application service 或 `platform/events`。
- Owner boundary:
  - `contest`：继续拥有公告、榜单更新、AWD 预览进度这些领域事实事件的发布。
  - `contest/infrastructure`：拥有 DB-backed realtime outbox 的持久化细节，并在可事务落库的写路径里和业务事务一起提交。
  - `ops/application`：把 contest 事件映射成 relay envelope，不直接知道 stream key、consumer group 或 reclaim 细节。
  - `ops/infrastructure`：拥有 Redis Streams key、消息编解码、dispatcher、consumer group、ACK/reclaim、DLQ 和本地 WebSocket fanout 细节。
  - `app/composition` / `ops/runtime`：拥有 dispatcher/consumer 的生命周期接线。
- Why this is the narrowest safe surface:
  - 只把 contest realtime relay 这一条跨实例短板做成可恢复链路，就能解决最先撞墙的事件丢失问题，同时不扩大到 notification、assessment 或整套事件总线重构。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这次不是单点 bugfix，而是在“保留单体边界”前提下重新选择 realtime relay 的物理落点，需要先压清 owner、非目标和跨实例语义。
- grill-with-docs findings:
  - `docs/design/backend-module-boundary-target.md` 已明确：`ops` 负责 WebSocket 广播适配，但不应成为业务模块的硬依赖；业务模块应该发布事件，`ops` 负责适配。
  - `docs/architecture/backend/07-modular-monolith-refactor.md` 已明确：`platform/events.Bus` 只负责进程内事件总线，不负责跨进程消息中间件语义。
  - `contest status transition` 已有 `事务内写 transition journal + 后台 replay side effects` 模式，说明仓库已经接受“持久化副作用账本 + 异步恢复”的结构，不需要发明第二套思路。
  - 当前前端事实源 `docs/architecture/frontend/05-websocket-composables.md` 已明确：contest announcements / scoreboard 仍是“轻事件 + HTTP 刷新”，因此第一刀不需要改 payload contract。
  - 活动 backlog 已记录 `assessment / ops` 的副作用事件化边界仍需收口；本次可先把 `ops` 的 contest realtime relay 物理落点与失败处理收口成单点 owner。
- Plan adjustments after challenge:
  - 不改 notification relay；避免把本次任务扩成“所有 WebSocket 全部外置化”。
  - 不新增前端 cursor 契约；客户端级恢复先明确排除，避免把实例级恢复和客户端级恢复混成一个任务。
  - Redis relay 从 `Pub/Sub` 升级成 `Streams`；stream key 和 per-instance cursor key 落到 `ops/infrastructure/cachekeys`。
  - dispatcher 和 stream consumer 都必须走 root background job 生命周期，不在 runtime/module 里私自悬挂 goroutine。
  - 对“可事务落库”的 realtime 事件，必须和业务提交一起写 outbox；不再接受“业务成功后顺手 best-effort 发事件”的伪可靠方案。

## Ordered Task Slices

### Slice 1: durable relay substrate

- 建表 `contest_realtime_outbox`，字段至少包含：事件类型、payload、aggregate 标识、dedupe key、status、attempt_count、next_attempt_at、last_error、stream_message_id、created_at、sent_at。
- 在 `contest` 侧增加 outbox entity/repository，复用当前 GORM repository 风格和 context contract。
- 在 `ops` 侧增加 stream adapter：
  - dispatcher：从 outbox 拉待发送记录，`XADD` 到 stream，成功后回写 outbox sent 状态和 stream id
  - consumer：按实例 cursor `XREAD` 补读，fanout 成功后推进 cursor
  - recovery：实例在 fanout 前退出时，因为 cursor 未推进，重启后仍会重新读到同一条 relay
  - retry：dispatcher 发布 stream 失败时增加 outbox attempt 并按退避时间重试
- 把 dispatcher / consumer 注册成 `ops` background jobs，纳入 root lifecycle。

### Slice 2: migrate contest realtime producers

- 先覆盖 DB-backed producer：
  - `announcement.created`
  - `announcement.deleted`
  - `scoreboard.updated`（仅手动 freeze/unfreeze 这类本来就带 PG 事务的管理链路）
- 将 `publishContestWeakEvent()` 从“直接调 bus 并吞错”改成“区分 durable producer 与 transient producer”，失败至少可见。
- 对 `submission_scoreboard_sync.go` 和 `awd_preview_realtime.go` 这类非事务型/高频链路，先接到 stream relay substrate，但在计划 review 中单列 producer durability 风险，不假装已经拥有事务级 outbox 保证。

### Slice 3: verification and convergence

- 增加 repository / stream tests 覆盖：
  - outbox 记录写入和状态迁移
  - consumer 推进 cursor 后不会回退重读
  - consumer 在推进 cursor 前中断时会重新读到同一条 relay
  - dispatcher 发布失败时会回写重试状态
- 跑最小充分 Go 验证。
- 进入独立 backend review，重点检查是否真的交付了“实例级恢复”，而不是只把消息路径从内存换成 Redis。

## Validation

- Commands:
  - `cd code/backend && go test ./internal/module/contest/infrastructure -run TestRepository.*RealtimeOutbox -count=1`
  - `cd code/backend && go test ./internal/module/ops/application/commands -run TestContestRealtimeService -count=1`
  - `cd code/backend && go test ./internal/module/ops/infrastructure -run TestContestRealtimeStream -count=1`
  - `cd code/backend && go test ./internal/module/contest/application/commands -run 'Test(PublishContestWeakEventDoesNotCreateBackgroundContext|ParticipationServiceCreateAnnouncementEnqueuesRealtimeRelay|SubmissionServiceSyncCorrectSubmissionScoreboardBroadcastsRealtimeEvent)' -count=1`
  - `cd code/backend && go test ./internal/app -run TestBuildOpsModuleDelegatesToContainerRuntime -count=1`
- Manual checks:
  - 确认 outbox dispatcher 崩溃重启后，未 sent 的记录会继续被推进 stream。
  - 确认实例在 fanout 前退出时，因为 cursor 未推进，重启后仍会补读到同一条 relay。
  - 确认 origin instance 不会同时走“直接本地 fanout + stream consumer fanout”导致重复推送。
  - 确认不同实例使用各自 cursor 追赶同一条 stream 时，announcement / scoreboard / AWD preview message 仍会各自落到本机 WebSocket 连接。
  - 确认 stream 发布失败时 outbox 会进入下一次重试，而不是完全沉默。
- Review focus:
  - `contest` 和 `ops` 是否按“contest 记 durable event 账本，ops 做 relay adapter”清楚分层。
  - dispatcher / consumer 的 goroutine、cancel、cursor 推进和重试是否纳入 root background job 生命周期。
  - 计划中是否明确区分了“producer durability”“instance consumption recovery”“client replay”三件事，没有混成一句泛泛的可靠性承诺。
  - 本次 touched surface 是否真的收口了 contest realtime relay 的结构债，而不是只把路径从内存换成 stream。
