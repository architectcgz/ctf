# contest realtime outbox worker must not revert sent rows on stale retries

## 问题描述

在多实例 contest realtime outbox dispatcher 场景里，多个实例会并发捞取同一批 pending 记录。如果 MarkRealtimeRelaySent / MarkRealtimeRelayFailed 只是按 id 无条件更新，较慢实例会把已被其他实例标记为 sent 的记录重新写回 pending，造成状态复活、重复重试，后续一旦 dedupe marker 丢失还会再次广播。已修正为 sent/failed 迁移都要求当前 status 仍为 pending；review 这类 DB outbox + 多实例 worker 组合时，应专门检查状态迁移是不是 compare-and-set，而不是只看发布侧幂等。

## 原因分析

- 这类问题只会在多实例或慢实例重试时暴露，单实例 happy path 很容易误判“实现已经幂等”。
- Redis publish 侧即使做了 dedupe marker，也只能保证“同一 dedupe key 不重复写 stream”；它并不能自动修复数据库里的 outbox 状态。
- 一旦 `sent -> pending` 被 stale worker 复活，系统后续就会持续把这条记录当成待发送任务；当 Redis marker 过期、被清理、被重建，或后续实现调整了 dedupe 策略时，就会重新把旧事件广播出去。
- review 如果只盯着“发布是否幂等”“消费者是否去重”，会漏掉最关键的 owner：outbox 行状态迁移本身是否具备 compare-and-set 语义。

## 解决方案

- `MarkRealtimeRelaySent` 和 `MarkRealtimeRelayFailed` 都改成条件更新：只有当前 `status = pending` 时才允许迁移。
- 较慢实例在拿到过期视图后，即使晚到执行 `MarkRealtimeRelayFailed`，也不会再把已经 `sent` 的记录写回 `pending`。
- 对 DB outbox + 多实例 worker 这类实现，review 时要单独检查三件事：
  - `ListPending...` 是否只是读取，不承担 claim 语义。
  - `sent/failed` 是否是条件迁移，而不是按 `id` 盲写。
  - 即使发布侧已有 Redis / MQ 幂等，DB outbox 状态是否仍然稳定、不会被 stale worker 复活。

## 收获

- 发布幂等和状态迁移幂等是两件事，不能因为前者存在就忽略后者。
- 多实例 worker 的正确性不能只看“最终有没有重复消息”，还要看“数据库里的任务状态会不会被过期执行流污染”。
- 对 outbox / inbox / saga record 这类 durable journal，最基础的保护不是重试次数，而是 compare-and-set 状态迁移。

## 沉淀状态

- 状态：implemented
- Owner：先沉淀在项目 `feedback/`，作为这次 realtime relay 外部化任务的 harness 复盘；后续如果同类问题再次出现，再上收到共享 backend / workflow review skill。
- 链接：
  - `/home/azhi/workspace/projects/ctf/code/backend/internal/module/contest/infrastructure/realtime_outbox_repository.go`
  - `/home/azhi/workspace/projects/ctf/code/backend/internal/module/ops/application/commands/contest_realtime_dispatcher.go`
  - `/home/azhi/workspace/projects/ctf/feedback/2026-06-07-contest-realtime-outbox-worker-must-not-revert-sent-rows-on-stale-retries.md`

## 证据

- file:
  - `/home/azhi/workspace/projects/ctf/code/backend/internal/module/contest/infrastructure/realtime_outbox_repository.go`
  - `/home/azhi/workspace/projects/ctf/code/backend/internal/module/contest/infrastructure/realtime_outbox_repository_test.go`
- command:
  - `go test ./internal/module/contest/infrastructure -run TestRealtimeOutboxRepository -count=1`
  - `go test ./internal/module/ops/application/commands -run 'TestContestRealtime(Service|OutboxDispatcher)' -count=1`
- behavior:
  - 先 `MarkRealtimeRelaySent`、后 `MarkRealtimeRelayFailed` 的 stale 执行流，不会再把已发送记录改回 `pending`。
  - 多实例 dispatcher 即使并发捞到同一条 pending 记录，也不会让 outbox 状态被晚到实例复活。

## Decision Log

- 2026-06-07: 在 contest realtime relay 外部化 review 中发现多实例 dispatcher 会无条件回写 outbox 状态。
- 2026-06-07: 代码已改为 `status = pending` 条件迁移，并补充仓储级回归测试。
- 2026-06-07: 该经验记录到项目 harness `feedback/`，用于后续 review 复用。
