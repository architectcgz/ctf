# Backend Review: Distributed Event Bus And Outbox Relay

## Review 对象

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-distributed-event-bus-and-outbox-relay`
- Branch: `task/2026-06-12-distributed-event-bus-and-outbox-relay`
- Diff source: current uncommitted diff
- Task slug: `2026-06-12-distributed-event-bus-and-outbox-relay`
- Implementation plan: `docs/plan/impl-plan/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md`
- Reviewer: independent `codex exec` read-only context

## 结论

- Classification check: 同意按非琐碎后端结构性任务处理，需要独立 review gate。
- Gate verdict: `pass`
- Material findings: 无当前未修复 material finding。
- History: 初审 4 个 blocker，第一次独立复审新增 1 个 blocker；5 个 blocker 均已在实现上下文修复并重跑相关验证。
- Final re-review: 独立只读复审确认五个 blocker 均已关闭，返回 `Gate verdict: pass`。
- Final composition re-review: 将 practice outbox handler 注册从 `router.go` 移到 composition root 后，独立只读复审确认 `router.go` 无 diff、内部 wiring 未重新打开 API router surface，返回 `Gate verdict: pass`。

## Findings

### Blocker 1: PostgreSQL partial unique index 与 GORM `ON CONFLICT` 不匹配

- Location:
  - `code/backend/internal/platform/events/outbox_repository.go:51`
  - `code/backend/migrations/000018_create_platform_event_outbox.up.sql:21`
- Risk:
  - 带 `DedupeKey` 的 outbox enqueue 在真实 PostgreSQL schema 下会失败。
- Reasoning:
  - migration 创建的是 partial unique index `WHERE dedupe_key <> ''`，但 GORM 当前生成 `ON CONFLICT (dedupe_key) DO NOTHING`，没有带 conflict target predicate。
  - 当前 `practice.flag_accepted`、`challenge.publish_check_finished`、`notification.created/read` 都设置非空 dedupe key。
  - 现有测试使用 SQLite `AutoMigrate`，没有覆盖真实 migration 形状。
- Required fix:
  - 让 schema 和 `ON CONFLICT` 对齐，并补 PostgreSQL/migration 级或 SQL shape 回归测试。

### Blocker 2: Challenge publish-check 发布状态与 outbox/job final 不在同一事务

- Location:
  - `code/backend/internal/module/challenge/application/challengepublishcheck/service.go:227`
  - `code/backend/internal/module/challenge/application/challengecore/service.go:275`
- Risk:
  - 可能出现“题目已发布、publish-check job 仍 running、通知事件丢失”。
- Reasoning:
  - `processPublishCheckJob` 先调用 `PublishChallenge` 写入 published 状态，随后才在另一个事务里更新 job final 并 enqueue outbox。
  - 若 outbox enqueue 失败，当前代码只记录 warning；失败测试只断言 job/outbox，没有覆盖 passed path 下 challenge 状态回滚。
- Required fix:
  - 把 challenge status update、publish-check job final update、outbox enqueue 放进同一个 tx owner。
  - weak catalog event 只能在事务成功后发，或后续迁入可恢复 outbox。
  - 补 passed self-check + outbox enqueue failure 时 challenge 仍为 draft 的测试。

### Blocker 3: Handler route 重试会重复创建通知

- Location:
  - `code/backend/internal/platform/events/outbox_dispatcher.go:40`
  - `code/backend/internal/module/ops/application/commands/notification_service.go:52`
- Risk:
  - handler route 的重复执行会无界创建重复通知。
- Reasoning:
  - dispatcher 顺序执行 handler，最后才 `MarkDispatched`。
  - `practice.flag_accepted` 注册 progress 删除和 notification 创建两个 handler；notification handler 内 `SendNotification` 每次插入新 `notifications` 行，并用新 notification ID 生成新的 fanout dedupe key。
  - `notifications` 表没有 source event key / handler execution key 约束。
- Required fix:
  - 给 handler side effect 增加源事件级幂等，例如 notification `source_event_key` partial unique，或 outbox handler execution 表。
  - 补 handler route retry / mark-dispatched failure 的重复通知测试。

### Blocker 4: `ClaimDue` 多 dispatcher 竞争时可绕过 backoff

- Location:
  - `code/backend/internal/platform/events/outbox_repository.go:73`
- Risk:
  - 失败事件可能被立即重新领取，backoff 失效并放大重复副作用。
- Reasoning:
  - `ClaimDue` 先 `Find` due rows，再逐条 `Update` lock；第二步只检查 `status` 和 `locked_until`，没有重新检查 `next_attempt_at <= now`。
  - 两个 worker 同时看到同一 due row 时，worker A dispatch 失败并 `MarkFailed(next_attempt_at=future)` 后，worker B 仍可能用旧 now 更新成功，因为 `locked_until` 已被清空。
- Required fix:
  - claim update 条件里重复 `next_attempt_at <= now`，或改为 PostgreSQL `FOR UPDATE SKIP LOCKED` 原子领取。
  - 补并发 claim/backoff 测试。

### Re-review Blocker 5: Publish-check finalization 会用 stale challenge 整行覆盖自检期间编辑

- Location:
  - `code/backend/internal/module/challenge/application/challengepublishcheck/service.go:248`
  - `code/backend/internal/module/challenge/infrastructure/repository.go:104`
- Risk:
  - 管理员在 self-check 运行期间修改题目时，publish-check 成功收尾会把 self-check 前读取的旧 `ChallengeWriteModel` 通过整行 `Save` 写回，覆盖 title、description、points 等字段。
- Reasoning:
  - 初次修复已把 challenge status、job final、outbox enqueue 放进同一事务，但事务内仍复用 self-check 前的 `challenge` 指针。
  - `Repository.Update` 使用 GORM `Save`，会把旧 model 的整行字段写回。
  - outbox payload 和 weak catalog event 也会基于旧 model 构造，导致事件事实和最终当前题目不一致。
- Required fix:
  - publish-check final transaction 内重新读取并锁定当前 challenge。
  - 发布状态更新只能定向更新 `status` / `updated_at`，不能复用整行 `Update`。
  - outbox payload 与 weak catalog event 使用 final transaction 内的当前 challenge fact。
  - 补回归测试：self-check 期间题目被编辑，publish-check 成功后编辑字段不能被覆盖。

## Fix Verification / Re-review Notes

- Blocker 1 fix:
  - `code/backend/internal/platform/events/outbox_repository.go` 的 `ON CONFLICT (dedupe_key)` 已补 `TargetWhere`，与 `code/backend/migrations/000018_create_platform_event_outbox.up.sql` 的 `WHERE dedupe_key <> ''` partial unique index 对齐。
  - 新增 / 更新回归覆盖：`TestOutboxRepositoryDedupeConflictTargetsPartialIndex`
- Blocker 2 fix:
  - `code/backend/internal/module/challenge/application/challengepublishcheck/service.go` 的 passed path 现在通过 `ChallengePublishCheckOutboxTxManager` 在同一个 transaction 内更新 challenge published 状态、publish-check job final 状态，并 enqueue `challenge.publish_check_finished` outbox。
  - weak catalog event 只在 transaction 成功后发送。
  - publish-check outbox enqueue failure 测试现在覆盖 passed path，并断言 challenge 仍为 draft、job 仍未 final。
- Blocker 3 fix:
  - `notifications.source_event_key` 已进入 migration、entity、repository port 和 repository 实现；handler-route notification 创建使用 `CreateIfSourceEventAbsent` 和 partial unique index 做源事件级幂等。
  - 新增 / 更新回归覆盖：`TestNotificationRepositoryCreateIfSourceEventAbsentIsIdempotent`、`TestNotificationServiceHandlePracticeFlagAcceptedOutboxEventIsIdempotent`
- Blocker 4 fix:
  - `OutboxRepository.ClaimDue` 的第二步 update 已重新检查 `next_attempt_at <= now`，stale read worker 不能绕过 future retry backoff。
  - 新增 / 更新回归覆盖：`TestOutboxRepositoryClaimDueDoesNotBypassFutureRetryAfterStaleRead`
- Re-review Blocker 5 fix:
  - `ChallengePublishCheckService` 不再依赖 core `ChallengePublisher` 执行发布，publish-check finalization 只通过 `ChallengePublishCheckOutboxTxManager` 完成。
  - `ChallengePublishCheckOutboxTxRepository` 新增 `LockChallengeByID` 和 `MarkChallengePublished`；final transaction 先锁定当前 challenge，再定向更新 `status` / `updated_at`。
  - outbox `challenge.publish_check_finished` payload 和 weak catalog event 都使用 final transaction 内读取到的当前 challenge fact。
  - TDD red: `cd code/backend && go test ./internal/module/challenge/application/commands -run TestServiceDispatchPublishCheckJobsDoesNotOverwriteChallengeEditedDuringSelfCheck -count=1` 按预期失败，显示 title / description / points 被旧 model 覆盖。
  - Regression: `TestServiceDispatchPublishCheckJobsDoesNotOverwriteChallengeEditedDuringSelfCheck`
- Post-fix validation evidence:
  - `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec|Dispatcher|StreamFanout' -count=1`: PASS
  - `cd code/backend && go test ./internal/module/ops/application/commands ./internal/module/ops/infrastructure ./internal/module/ops/api/http -run 'Notification|Outbox|Fanout' -count=1`: PASS
  - `cd code/backend && go test ./internal/module/challenge/... -run 'PublishCheck|Outbox' -count=1`: PASS
  - `cd code/backend && go test ./internal/platform/events ./internal/module/ops/... ./internal/module/practice/... ./internal/module/challenge/... -run 'Outbox|Stream|Notification|FlagAccepted|PublishCheck|Progress' -count=1`: PASS
  - `bash scripts/check-architecture.sh --full`: PASS
  - `python3 scripts/check-docs-consistency.py`: PASS
  - `git diff --check -- docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/impl-plan/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/testutil/systemapp code/backend/internal/app code/backend/migrations`: PASS
- Re-review focus:
  - 复核上述五个 blocker 的修复是否完整。
- Final independent re-review:
  - Command: `codex exec --sandbox read-only --cd /home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-distributed-event-bus-and-outbox-relay --output-last-message /tmp/t4-outbox-rereview-after-fix.md`
  - Result: `Gate verdict: pass`
  - Findings: no material findings.
  - Reviewer check: `git diff --check -- docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/impl-plan/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/testutil/systemapp code/backend/internal/app code/backend/migrations`: PASS
- Final composition-root re-review:
  - Command: `codex exec --sandbox read-only --cd /home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-distributed-event-bus-and-outbox-relay --output-last-message /tmp/t4-outbox-final-rereview.md`
  - Result: `Gate verdict: pass`
  - Findings: no material findings.
  - Notes: 复审确认 `BuildPracticeModule` 先注册 progress handler，再执行 root 中排队的 ops notification registrar；`router.go` 当前无 diff，内部 outbox handler 接线不再落在 API router surface。

## 必须重跑的验证

- `cd code/backend && go test ./internal/platform/events -run 'Outbox|Codec|Dispatcher|StreamFanout' -count=1`
- `cd code/backend && go test ./internal/module/ops/application/commands ./internal/module/ops/infrastructure ./internal/module/ops/api/http -run 'Notification|Outbox|Fanout' -count=1`
- `cd code/backend && go test ./internal/module/challenge/... -run 'PublishCheck|Outbox' -count=1`
- `cd code/backend && go test ./internal/platform/events ./internal/module/ops/... ./internal/module/practice/... ./internal/module/challenge/... -run 'Outbox|Stream|Notification|FlagAccepted|PublishCheck|Progress' -count=1`
- `bash scripts/check-architecture.sh --full`
- `python3 scripts/check-docs-consistency.py`
- `git diff --check -- docs/reviews/backend/2026-06-12-backend-review-distributed-event-bus-and-outbox-relay.md docs/architecture/backend/01-system-architecture.md docs/architecture/backend/05-key-flows.md docs/plan/impl-plan/2026-06-12-distributed-event-bus-and-outbox-relay-implementation-plan.md code/backend/internal/platform/events code/backend/internal/module/ops code/backend/internal/module/practice code/backend/internal/module/challenge code/backend/internal/testutil/systemapp code/backend/internal/app code/backend/migrations`

## 残余风险

- 真实多副本 WebSocket fanout 仍缺环境演练证据；该项属于后续集成验收风险，不是本次 code-review gate blocker。
- stream consumer cursor identity 当前基于 hostname；一宿主机 / 一 pod 一个 API 进程的部署形态下可用，如果同一 hostname 下运行多个 API 进程，需要引入更细的 instance id。
- parent input `docs/plan/impl-plan/2026-06-12-true-ha-group/distributed-event-bus-and-outbox-relay.md` 在当前 worktree 不存在，本次 review 以当前 implementation plan、后端架构文档和源码 diff 为依据。

## Touched Known-Debt Status

- `practice` submission + outbox tx owner 在 touched surface 内收口。
- `challenge` publish-check tx owner 初审时未收口，第一次复审又发现 stale full-row overwrite 风险；当前实现已改为 tx 内锁定当前 challenge + 定向发布状态更新，并经独立复审确认。
- `ops/runtime/module.go` 拆分为 `notification_wiring.go` / `contest_realtime_wiring.go` 是正向收口，未发现该拆分本身引入 owner 混杂 blocker。
- outbox/dispatcher retry/idempotency 是本任务核心债务；实现上下文已补 partial unique conflict target、ClaimDue backoff recheck 和 notification source-event idempotency，并经独立复审确认。
