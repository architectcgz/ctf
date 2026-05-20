# Contest Status State Machine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将竞赛状态推进从“扫描后直接写状态”升级为严格幂等的状态机，并保留多 `ctf-api` 实例下的分布式锁排他能力。

**Architecture:** 状态机正确性由领域迁移校验、数据库 compare-and-set 条件更新和副作用幂等保证；`contest:status_updater:lock` 只承担多实例调度排他、降噪和减负载。第一阶段不新增表，先把自动状态推进改为 `applied/stale/invalid/not_found` 语义；后续阶段再引入统一迁移服务、迁移记录表和 outbox 扩展点。

**Tech Stack:** Go 1.24, GORM, PostgreSQL, Redis, miniredis, zap, existing `redislock`.

---

## Plan Summary

### Objective

完成 `docs/architecture/backend/design/contest-status-state-machine.md` 中定义的三个阶段：

- 第一阶段：保留分布式锁，补 `contest_status_updater` 锁续租，将自动状态推进改为条件迁移，并且只在 `applied=true` 后执行副作用。
- 第二阶段：引入 `ContestStatusTransitionService`，收口自动调度和人工管理入口的状态迁移校验。
- 第三阶段：增加迁移记录表和 outbox 扩展点，让副作用具备审计、重试和去重基础。

### Non-goals

- 不改变对外 API 的状态枚举。
- 不重构竞赛 CRUD、排行榜、AWD round updater 或前端页面。
- 不一次性迁移所有历史状态数据；本计划只处理后续状态推进语义。
- 不把分布式锁删除；锁仍然作为多实例调度排他层保留。

### Source architecture or design docs

- `docs/architecture/backend/design/contest-status-state-machine.md`
- `docs/architecture/backend/design/README.md`
- `AGENTS.md`

### Current implementation evidence

- 状态枚举：`code/backend/internal/model/contest.go`
- 合法迁移表：`code/backend/internal/module/contest/domain/contest.go`
- 自动状态调度器：`code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- 状态副作用：`code/backend/internal/module/contest/application/jobs/status_update_support.go`
- 状态仓储接口：`code/backend/internal/module/contest/ports/contest.go`
- 当前直接更新实现：`code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go`
- 调度器测试：`code/backend/internal/module/contest/application/jobs/status_updater_test.go`
- 配置默认值：`code/backend/internal/config/config.go`

### Architecture Evaluation

评估结论：

- 正确性边界必须落到状态机和数据库条件更新，不能只依赖 Redis 分布式锁。
- 副作用必须跟随“迁移成功”而不是跟随“内存中计算出新状态”。
- `StatusUpdater` 可以继续负责定时扫描，但不应长期拥有状态迁移提交细节。
- 第三阶段的迁移记录表不应成为第一阶段正确性的前置条件，否则会把一个本可小步验证的修复变成大迁移。
- outbox 只在副作用扩展到异步消息、通知、审计投递或外部系统时真正启用；本计划先留出结构和测试边界，不为当前 Redis 副作用强行引入完整发布链路。

### Dependency order

1. 先补 `redislock.Refresh` 和状态调度器锁续租，消除长任务 TTL 过期问题。
2. 再定义状态迁移结果类型和条件更新仓储方法。
3. 调整 `StatusUpdater`：先 apply 迁移，再执行副作用。
4. 引入统一 `ContestStatusTransitionService`，让自动调度和人工状态变更复用同一校验入口。
5. 增加迁移记录表，记录成功迁移和副作用结果。
6. 预留 outbox 事件结构，并把副作用失败记录为可重试状态。
7. 完成集成验证、并发验证和文档同步。

### Expected specialist skills

- `backend-engineer`
- `go-backend`
- `test-engineer`
- `doc-admin-agent`
- `code-reviewer`

## File Structure

### Domain and application

- Modify: `code/backend/internal/module/contest/domain/contest.go`
  - 保留 `IsValidTransition`，新增自动迁移判定 helper 或 typed transition validation。
- Create: `code/backend/internal/module/contest/domain/status_transition.go`
  - 定义 `ContestStatusTransition`、`ContestStatusTransitionResult`、`ContestStatusTransitionOutcome`。
- Create: `code/backend/internal/module/contest/application/jobs/status_transition_service.go`
  - 实现状态迁移服务，负责校验、调用仓储条件更新、生成结果。
- Modify: `code/backend/internal/module/contest/application/jobs/status_updater.go`
  - 注入迁移服务或在构造函数中默认创建。
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
  - 使用续租锁上下文；调用迁移服务；仅 `Applied` 时执行副作用。
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_support.go`
  - 副作用入口改为接收 transition result；关键实现点加注释。
- Modify: `code/backend/internal/module/contest/application/commands/contest_update_support.go`
  - 人工状态变更复用状态机校验；第二阶段再接入迁移服务。

### Ports and infrastructure

- Modify: `code/backend/internal/module/contest/ports/contest.go`
  - 将 `ContestStatusRepository` 的写方法从 `UpdateStatus` 调整为条件迁移方法。
- Modify: `code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go`
  - 实现 `ApplyStatusTransition(ctx, transition)`，使用 `WHERE id = ? AND status = ?`。
- Create: `code/backend/internal/module/contest/infrastructure/contest_status_transition_repository.go`
  - 第三阶段实现迁移记录写入和副作用状态更新。

### Redis lock reuse

- Modify: `code/backend/internal/pkg/redislock/lock.go`
  - 新增 token 校验续租 `Refresh(ctx, ttl)`。
- Create: `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
  - 提取 status updater 和 AWD round updater 可复用的持锁续租 helper。
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`
  - 如果当前分支已有 AWD 续租实现，将其改为复用 helper；否则本任务只新增 helper 并接入 status updater。

### Migrations and outbox

- Create: `code/backend/internal/database/migrations/*_create_contest_status_transitions.*`
  - 具体目录和文件名以现有 migration 体系为准。
- Optional Create: `code/backend/internal/database/migrations/*_create_contest_status_transition_outbox.*`
  - 仅当项目已有 outbox 模式或本阶段决定落库 outbox 时创建。

### Tests

- Modify: `code/backend/internal/module/contest/application/jobs/status_updater_test.go`
- Create: `code/backend/internal/module/contest/application/jobs/status_transition_service_test.go`
- Create: `code/backend/internal/module/contest/infrastructure/contest_status_update_repository_test.go`
- Create: `code/backend/internal/module/contest/application/jobs/lock_keepalive_test.go`
- Modify as needed: `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- Modify as needed: `code/backend/internal/config/config_test.go` only if config keys change. This plan should not require config changes.

### Docs

- Modify: `docs/architecture/backend/design/contest-status-state-machine.md`
- Modify: `docs/architecture/backend/design/README.md` only if status changes from “设计中” to “已采用” after implementation.

## Task 1: Baseline and Lock Keepalive Foundation

**Goal:** 先解决 `contest_status_update_lock_expired_before_release` 的同类锁 TTL 问题，并把续租能力做成后续可复用基础。

**Files:**
- Modify: `code/backend/internal/pkg/redislock/lock.go`
- Create: `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
- Test: `code/backend/internal/module/contest/application/jobs/lock_keepalive_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_runner.go`

- [ ] **Step 1: Write failing test for token-checked refresh**

Create or extend a focused Redis lock test. If `redislock` package has no test file, create:

`code/backend/internal/pkg/redislock/lock_test.go`

Test shape:

```go
func TestLockRefreshExtendsOwnedLock(t *testing.T) {
    // acquire lock with a short TTL
    // wait less than TTL
    // call Refresh
    // assert key still exists after original TTL would have expired
}

func TestLockRefreshDoesNotExtendForeignLock(t *testing.T) {
    // acquire lock
    // replace key token manually
    // call Refresh
    // expect refreshed=false
}
```

- [ ] **Step 2: Run refresh tests and verify failure**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/pkg/redislock -run 'TestLockRefresh' -count=1
```

Expected: FAIL because `Refresh` does not exist.

- [ ] **Step 3: Implement `redislock.Lock.Refresh`**

Add token-checked Lua refresh:

```go
const refreshScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("pexpire", KEYS[1], ARGV[2])
else
    return 0
end
`
```

Implementation requirement:

- Use `PEXPIRE`.
- Convert TTL to milliseconds.
- Return `(false, nil)` for nil lock, empty key, empty token, non-positive TTL.
- Add a concise comment explaining why token check is required.

- [ ] **Step 4: Run redislock tests**

Run:

```bash
go test ./internal/pkg/redislock -count=1
```

Expected: PASS.

- [ ] **Step 5: Write failing keepalive test**

Create `code/backend/internal/module/contest/application/jobs/lock_keepalive_test.go`.

Test cases:

- keepalive refreshes a lock while a simulated job sleeps past original TTL.
- keepalive cancels the run context when `Refresh` returns false.

Use miniredis and a short TTL, for example `60ms`.

- [ ] **Step 6: Implement shared keepalive helper**

Create `lock_keepalive.go` with a small helper owned by jobs package:

```go
type lockKeepaliveConfig struct {
    Name string
    TTL  time.Duration
}

func startRedisLockKeepalive(ctx context.Context, log *zap.Logger, lock *redislock.Lock, cfg lockKeepaliveConfig) (context.Context, func())
```

Implementation requirements:

- Derive a cancelable run context from input `ctx`.
- Refresh every `ttl/3`, or `ttl/2` for very small TTLs.
- If refresh returns false, log a warn and cancel run context.
- If refresh errors, log error and continue until context ends.
- Stop function must stop ticker goroutine and wait for it.
- Add comments at the key points: why keepalive exists, why false refresh cancels the run, why stop waits.

- [ ] **Step 7: Run keepalive tests**

Run:

```bash
go test ./internal/module/contest/application/jobs -run 'TestRedisLockKeepalive' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add code/backend/internal/pkg/redislock/lock.go \
  code/backend/internal/pkg/redislock/lock_test.go \
  code/backend/internal/module/contest/application/jobs/lock_keepalive.go \
  code/backend/internal/module/contest/application/jobs/lock_keepalive_test.go
git commit -m "feat(竞赛): 增加分布式锁续租能力"
```

## Task 2: Apply Keepalive to Contest Status Updater

**Goal:** `StatusUpdater` 持锁期间自动续租，释放锁不被上层取消打断。

**Files:**
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- Test: `code/backend/internal/module/contest/application/jobs/status_updater_test.go`

- [ ] **Step 1: Write failing status updater lock test**

Add a test similar to:

```go
func TestStatusUpdaterRefreshesSchedulerLockWhileRunning(t *testing.T) {
    // use miniredis
    // repo ListByStatusesAndTimeRange blocks or sleeps longer than lock TTL
    // assert lock key still exists during run
    // assert key is released after run
}
```

Expected behavior:

- During the long run, `contest:status_updater:lock` still has positive TTL.
- After run exits, lock key is released.

- [ ] **Step 2: Run test and verify failure**

Run:

```bash
go test ./internal/module/contest/application/jobs -run 'TestStatusUpdaterRefreshesSchedulerLockWhileRunning' -count=1
```

Expected: FAIL with lock missing or expired.

- [ ] **Step 3: Update `status_update_runner.go` to use keepalive**

Implementation requirements:

- After lock acquired, call shared keepalive helper.
- Use returned `runCtx` for repository query and per-contest processing.
- Check `runCtx.Err()` before each contest.
- Release lock with `context.WithoutCancel(ctx)` and short timeout.
- Add comments explaining long-running scheduler lease renewal and release context behavior.

- [ ] **Step 4: Run focused status updater tests**

Run:

```bash
go test ./internal/module/contest/application/jobs -run 'TestStatusUpdater' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add code/backend/internal/module/contest/application/jobs/status_update_runner.go \
  code/backend/internal/module/contest/application/jobs/status_updater_test.go
git commit -m "fix(竞赛): 为状态调度锁增加续租"
```

## Task 3: Introduce Conditional Transition Port

**Goal:** 将状态写入从 `UpdateStatus(id, status)` 改为显式 `from -> to` 条件迁移。

**Files:**
- Modify: `code/backend/internal/module/contest/ports/contest.go`
- Modify: `code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go`
- Create: `code/backend/internal/module/contest/infrastructure/contest_status_update_repository_test.go`

- [ ] **Step 1: Define transition DTOs in domain package**

Create `code/backend/internal/module/contest/domain/status_transition.go`:

```go
type ContestStatusTransition struct {
    ContestID  int64
    FromStatus string
    ToStatus   string
    Reason     string
    OccurredAt time.Time
}

type ContestStatusTransitionResult struct {
    Transition    ContestStatusTransition
    Applied       bool
    StatusVersion int64
}
```

Reason: ports and infrastructure both need the type, so it must not live in `application/jobs`.

- [ ] **Step 2: Change `ContestStatusRepository` interface**

Replace:

```go
UpdateStatus(ctx context.Context, id int64, status string) error
```

With:

```go
ApplyStatusTransition(ctx context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error)
```

Exact package can be adjusted to avoid cycles, but the method must include `fromStatus` and `toStatus`.

- [ ] **Step 3: Write failing repository tests**

Test cases:

- `registration -> running` with current `registration` returns `Applied=true`.
- Same transition repeated returns `Applied=false`.
- Missing contest returns `ErrContestNotFound`.
- `running -> frozen` does not update if DB currently has `ended`.

- [ ] **Step 4: Implement conditional update**

Repository update must use:

```go
WHERE id = ? AND status = ? AND deleted_at IS NULL
```

Rows affected handling:

- `RowsAffected == 1`: return `Applied=true`.
- `RowsAffected == 0` and contest missing: return `ErrContestNotFound`.
- `RowsAffected == 0` and contest exists: return `Applied=false`, nil error.

- [ ] **Step 5: Run repository tests**

Run:

```bash
go test ./internal/module/contest/infrastructure -run 'TestRepositoryApplyStatusTransition' -count=1
```

Expected: PASS.

- [ ] **Step 6: Fix compile errors in tests/stubs**

Update `statusUpdaterRepoStub` in `status_updater_test.go` to implement `ApplyStatusTransition`.

- [ ] **Step 7: Run contest packages compile check**

Run:

```bash
go test ./internal/module/contest/... -count=1
```

Expected: PASS or only failures unrelated to this change. Any compile error from the interface change must be fixed in this task.

- [ ] **Step 8: Commit**

```bash
git add code/backend/internal/module/contest/ports/contest.go \
  code/backend/internal/module/contest/domain/status_transition.go \
  code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go \
  code/backend/internal/module/contest/infrastructure/contest_status_update_repository_test.go \
  code/backend/internal/module/contest/application/jobs/status_updater_test.go
git commit -m "feat(竞赛): 使用条件迁移更新比赛状态"
```

## Task 4: Bind Side Effects to Applied Transitions

**Goal:** 只有成功提交状态迁移的实例才能执行封榜快照和结束清理。

**Files:**
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_updater_test.go`

- [ ] **Step 1: Write failing stale transition side-effect test**

Add test:

```go
func TestStatusUpdaterSkipsSideEffectsWhenTransitionIsStale(t *testing.T)
```

Set repo stub to return `Applied=false` for `running -> frozen`.

Expected:

- `createFrozenSnapshot` equivalent is not called.
- No `contest_status_updated` success path assumption is made.

If direct function spying is awkward, use Redis:

- Seed ranking source key.
- Stale transition should not create frozen snapshot key.

- [ ] **Step 2: Write failing applied transition side-effect tests**

Cases:

- `running -> frozen` with `Applied=true` creates frozen snapshot.
- `running -> ended` with `Applied=true` clears AWD runtime state.
- `frozen -> ended` with `Applied=true` clears AWD runtime state.

- [ ] **Step 3: Update runner ordering**

Current order creates frozen snapshot before status update. Change order to:

1. calculate next status
2. block AWD automatic start if readiness not ready
3. validate and apply transition
4. if not applied, continue
5. execute side effects for the applied transition
6. log success

Add a concise comment around step 4/5 explaining that side effects belong to the winner of the compare-and-set transition.

- [ ] **Step 4: Run status updater tests**

Run:

```bash
go test ./internal/module/contest/application/jobs -run 'TestStatusUpdater' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add code/backend/internal/module/contest/application/jobs/status_update_runner.go \
  code/backend/internal/module/contest/application/jobs/status_update_support.go \
  code/backend/internal/module/contest/application/jobs/status_updater_test.go
git commit -m "fix(竞赛): 仅在状态迁移成功后执行副作用"
```

## Task 5: Add Transition Validation Service

**Goal:** 将状态合法性校验从调度器细节中抽出来，形成可复用的状态迁移服务。

**Files:**
- Create: `code/backend/internal/module/contest/application/jobs/status_transition_service.go`
- Create: `code/backend/internal/module/contest/application/jobs/status_transition_service_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_updater.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_runner.go`

- [ ] **Step 1: Write service tests for valid and invalid transitions**

Test cases:

- `registration -> running` valid and applied.
- `running -> frozen` valid and applied.
- `frozen -> ended` valid and applied.
- `registration -> ended` invalid.
- `ended -> running` invalid.
- stale repository result returns `Applied=false` without error.

- [ ] **Step 2: Implement service**

Service responsibilities:

- Reject empty contest ID, empty `from` or `to`.
- Reject no-op transitions.
- Call `domain.IsValidTransition(from, to)`.
- Call repository `ApplyStatusTransition`.
- Return structured result.

Do not put Redis side effects inside this service.

- [ ] **Step 3: Wire `StatusUpdater` through service**

Options:

- Preferred: `StatusUpdater` stores `transitioner contestStatusTransitioner`.
- Constructor creates default service from repo.
- Tests can inject stub service if needed.

- [ ] **Step 4: Run service and updater tests**

Run:

```bash
go test ./internal/module/contest/application/jobs -run 'TestStatusTransition|TestStatusUpdater' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add code/backend/internal/module/contest/application/jobs/status_transition_service.go \
  code/backend/internal/module/contest/application/jobs/status_transition_service_test.go \
  code/backend/internal/module/contest/application/jobs/status_updater.go \
  code/backend/internal/module/contest/application/jobs/status_update_runner.go
git commit -m "feat(竞赛): 收敛比赛状态迁移服务"
```

## Task 6: Reuse State Machine in Manual Contest Status Updates

**Goal:** 人工管理入口不再绕过状态机直接写字符串状态。

**Files:**
- Modify: `code/backend/internal/module/contest/application/commands/contest_update_support.go`
- Modify as needed: `code/backend/internal/module/contest/application/commands/contest_service.go`
- Modify as needed: `code/backend/internal/module/contest/ports/contest.go`
- Test: existing contest command tests under `code/backend/internal/module/contest/application/commands`

- [ ] **Step 1: Inspect manual update flow**

Read:

```bash
sed -n '1,180p' code/backend/internal/module/contest/application/commands/contest_update_support.go
rg -n "Update\\(ctx.*Contest|Status" code/backend/internal/module/contest/application/commands -S
```

Expected: identify exact method that applies admin status changes.

- [ ] **Step 2: Write or update tests**

Test cases:

- valid manual `draft -> registration` still works.
- invalid manual transition still returns existing validation error.
- manual status update does not bypass `domain.IsValidTransition`.

- [ ] **Step 3: Route manual status validation through shared service or shared domain validator**

Minimum acceptable implementation:

- Manual path calls the same `domain.IsValidTransition`.
- It does not need to use the scheduler transition service if that would force unrelated CRUD repository changes.

Preferred implementation:

- Introduce a command-level transition service wrapper if manual status updates need compare-and-set semantics.

- [ ] **Step 4: Run command tests**

Run:

```bash
go test ./internal/module/contest/application/commands -run 'Test.*Contest.*Status|Test.*Update' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add code/backend/internal/module/contest/application/commands \
  code/backend/internal/module/contest/ports/contest.go
git commit -m "refactor(竞赛): 统一人工状态迁移校验"
```

## Task 7: Add Transition Record Persistence

**Goal:** 为状态迁移增加可审计、可重试的持久化记录，并用 `status_version` 支持同一状态对在未来被合法重复消费。

**Files:**
- Create: migration under existing migration directory
- Modify: `code/backend/internal/model` or equivalent model location
- Modify: `code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go`
- Create: `code/backend/internal/module/contest/infrastructure/contest_status_transition_repository.go`
- Test: `code/backend/internal/module/contest/infrastructure/contest_status_transition_repository_test.go`

- [ ] **Step 1: Locate migration convention**

Run:

```bash
find code/backend -maxdepth 4 -type d | rg 'migration|migrations'
find code/backend -maxdepth 5 -type f | rg 'migration|migrations' | head -n 40
```

Expected: exact migration directory and naming convention identified.

- [ ] **Step 2: Write migration**

Add `status_version` to `contests`:

```text
status_version bigint NOT NULL DEFAULT 0
```

Create `contest_status_transitions` with:

```text
id
contest_id
status_version
from_status
to_status
reason
occurred_at
applied_by
side_effect_status
side_effect_error
created_at
updated_at
```

Indexes:

- `contest_id`
- `(contest_id, status_version)` unique
- `occurred_at`

Do not use `(contest_id, from_status, to_status)` as the unique key. That would block a future legal flow such as `registration -> draft -> registration -> running`.

- [ ] **Step 3: Add model**

Add a model struct matching existing GORM conventions.

Required fields:

- contest status version
- status strings
- reason
- side effect status enum string
- nullable side effect error
- timestamps in UTC.

- [ ] **Step 4: Write repository tests**

Test cases:

- create transition record on applied transition.
- duplicate `(contest_id, status_version)` is handled as already recorded, not fatal.
- update side effect status to `succeeded`.
- update side effect status to `failed` with error message.

- [ ] **Step 5: Update conditional transition to increment version**

Update `ApplyStatusTransition` so an applied transition increments `contests.status_version`.

Implementation options:

- Preferred on PostgreSQL: use `RETURNING status_version` to return the new version.
- If GORM test dialect makes `RETURNING` awkward, perform the update in a transaction and reload the row after `RowsAffected == 1`.

The returned `ContestStatusTransitionResult` must include the resulting `StatusVersion`.

- [ ] **Step 6: Implement transition record repository**

Keep repository small:

- `RecordAppliedTransition(ctx, result)`
- `MarkTransitionSideEffectsSucceeded(ctx, transitionID)`
- `MarkTransitionSideEffectsFailed(ctx, transitionID, err)`

- [ ] **Step 7: Run migration/model repository tests**

Run:

```bash
go test ./internal/module/contest/infrastructure -run 'Test.*StatusTransition' -count=1
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add code/backend/internal/database \
  code/backend/internal/model \
  code/backend/internal/module/contest/domain/status_transition.go \
  code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go \
  code/backend/internal/module/contest/infrastructure/contest_status_transition_repository.go \
  code/backend/internal/module/contest/infrastructure/contest_status_transition_repository_test.go
git commit -m "feat(竞赛): 记录比赛状态迁移"
```

## Task 8: Wire Transition Records and Side Effect Status

**Goal:** 将迁移记录接入状态迁移服务，并在副作用完成后更新记录状态。

**Files:**
- Modify: `code/backend/internal/module/contest/application/jobs/status_transition_service.go`
- Modify: `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
- Modify: `code/backend/internal/module/contest/runtime/module.go`
- Test: `code/backend/internal/module/contest/application/jobs/status_transition_service_test.go`
- Test: `code/backend/internal/module/contest/application/jobs/status_updater_test.go`

- [ ] **Step 1: Extend interfaces**

Add small consumer-owned interfaces:

```go
type contestStatusTransitionRecorder interface {
    RecordAppliedTransition(ctx context.Context, result ContestStatusTransitionResult) (int64, error)
    MarkTransitionSideEffectsSucceeded(ctx context.Context, id int64) error
    MarkTransitionSideEffectsFailed(ctx context.Context, id int64, cause error) error
}
```

- [ ] **Step 2: Write tests for record creation**

Cases:

- Applied transition creates record.
- Stale transition does not create record.
- Record failure logs/returns error according to chosen policy.

Preferred policy:

- If state update applied but record insert fails, log error and continue side effects, because state is already committed.
- Mark as risk in logs; do not roll back status.

- [ ] **Step 3: Update side effect status after execution**

After side effects:

- if all side effects succeed: mark succeeded.
- if any side effect fails: mark failed and keep state as migrated.

This may require changing side effect helpers to return error instead of only logging.

- [ ] **Step 4: Wire repository in runtime module**

`contestinfra.Repository` can implement both status update and record methods if it owns DB.

Update constructor wiring in:

`code/backend/internal/module/contest/runtime/module.go`

- [ ] **Step 5: Run job and runtime tests**

Run:

```bash
go test ./internal/module/contest/application/jobs ./internal/module/contest/runtime -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add code/backend/internal/module/contest/application/jobs \
  code/backend/internal/module/contest/infrastructure \
  code/backend/internal/module/contest/runtime/module.go
git commit -m "feat(竞赛): 记录状态迁移副作用结果"
```

## Task 9: Add Outbox Extension Point

**Goal:** 为未来通知、审计投递或外部副作用提供 outbox 扩展点，不把当前 Redis 副作用强行改成异步。

**Files:**
- Create or Modify: outbox-related files only after checking existing project patterns
- Modify: `docs/architecture/backend/design/contest-status-state-machine.md`
- Test: focused outbox tests if a table or repository is added

- [ ] **Step 1: Inspect existing outbox/event patterns**

Run:

```bash
rg -n "outbox|event|publisher|Publish|Inbox|dedupe" code/backend/internal -S
```

Expected: determine whether project already has an outbox convention.

- [ ] **Step 2: Choose implementation level**

Decision rule:

- If existing outbox infrastructure exists, add `contest_status_transition_applied` event writing.
- If no outbox infrastructure exists, document extension point and do not add a standalone half-pattern unless a concrete async consumer is also implemented.

- [ ] **Step 3: If implementing outbox, write tests first**

Test:

- applied transition writes one outbox event.
- stale transition writes no outbox event.
- duplicate transition does not duplicate event.

- [ ] **Step 4: Implement or document deferral**

If implemented:

- Add migration/table/model according to existing convention.
- Write event payload with `transition_id`, `contest_id`, `from_status`, `to_status`, `occurred_at`.

If deferred:

- Update architecture doc with explicit reason: current side effects are Redis-local and idempotent; full outbox waits until external async side effects exist.

- [ ] **Step 5: Run tests or doc check**

Run relevant tests if code changed:

```bash
go test ./internal/module/contest/... -count=1
```

If doc-only:

```bash
git diff --check docs/architecture/backend/design/contest-status-state-machine.md
```

- [ ] **Step 6: Commit**

```bash
git add docs/architecture/backend/design/contest-status-state-machine.md code/backend/internal
git commit -m "feat(竞赛): 预留状态迁移 outbox 扩展点"
```

## Task 10: Integration Verification and Architecture Docs

**Goal:** 验证所有阶段组合后的行为，并同步架构文档状态。

**Files:**
- Modify: `docs/architecture/backend/design/contest-status-state-machine.md`
- Modify: `docs/architecture/backend/design/README.md`
- Test: backend contest packages and app-level tests

- [ ] **Step 1: Run focused package tests**

Run:

```bash
cd /home/azhi/workspace/projects/ctf/code/backend
go test ./internal/pkg/redislock \
  ./internal/module/contest/domain \
  ./internal/module/contest/infrastructure \
  ./internal/module/contest/application/jobs \
  ./internal/module/contest/application/commands \
  -count=1
```

Expected: PASS.

- [ ] **Step 2: Run broader contest tests**

Run:

```bash
go test ./internal/module/contest/... -count=1
```

Expected: PASS.

- [ ] **Step 3: Run app routing/state matrix tests if impacted**

Run:

```bash
go test ./internal/app -run 'TestFullRouterStateMatrix|TestRouter' -count=1
```

Expected: PASS, or document unrelated existing failures with exact error.

- [ ] **Step 4: Run migration verification**

Use the repo's existing migration test or startup test. If no explicit migration command exists, run the smallest backend integration test that initializes DB migrations.

Expected: `contest_status_transitions` migration applies cleanly.

- [ ] **Step 5: Update architecture doc status**

In `docs/architecture/backend/design/contest-status-state-machine.md`:

- Change status from `设计中` to `已采用` only after implementation and tests pass.
- Add a short “落地状态” section listing implemented stages and any deferred outbox decision.

- [ ] **Step 6: Self-review correctness**

Review checklist:

- No side effect runs before `Applied=true`.
- `Applied=false` is not logged as an error.
- `ended` cannot migrate out through automatic scheduler.
- `StatusUpdater` uses run context returned by lock keepalive.
- Lock release uses short timeout and `context.WithoutCancel`.
- Key implementation points have comments explaining why, especially lock renewal and side-effect ownership.

- [ ] **Step 7: Commit docs and final fixes**

```bash
git add docs/architecture/backend/design/contest-status-state-machine.md \
  docs/architecture/backend/design/README.md \
  code/backend
git commit -m "docs(竞赛): 更新状态机落地说明"
```

## Final Verification Checklist

- [ ] `go test ./internal/pkg/redislock -count=1`
- [ ] `go test ./internal/module/contest/... -count=1`
- [ ] `go test ./internal/app -run 'TestFullRouterStateMatrix|TestRouter' -count=1`
- [ ] Migration verification completed or explicitly documented.
- [ ] Architecture doc updated from design to implementation status.
- [ ] No unrelated working tree changes included in commits.

## Risk Notes

- `contest_status_transitions` 的唯一约束应使用 `(contest_id, status_version)`。不要使用 `(contest_id, from_status, to_status)`，因为人工回退后再次发布可能合法重复同一状态对。
- 如果当前分支已经包含 AWD scheduler 锁续租改动，Task 1 需要避免重复实现，直接抽取复用 helper。
- 如果项目没有 outbox 体系，Task 9 应以文档化扩展点收尾，不应引入没有 consumer 的半套异步框架。
- 副作用失败后的重试策略需要保持保守：状态不回滚，记录失败，后续由调度重跑或专门修复任务补偿。
