# 竞赛状态机设计

> 状态：Current
> 事实源：`code/backend/internal/model/contest.go`、`code/backend/internal/module/contest/application/jobs/`、`code/backend/internal/module/contest/infrastructure/`
> 替代：无

## 定位

本文档说明当前竞赛状态机的状态集合、推进方式和副作用重放边界。

- 负责：描述 `contests.status`、`contest_status_transitions`、调度锁与手动 / 自动状态迁移的统一口径。
- 不负责：继续把旧的 `published / cancelled / archived` 等历史候选状态写成当前后端事实。

## 当前设计

- `code/backend/internal/model/contest.go`、`code/backend/internal/module/contest/domain/status_transition.go`
  - 负责：定义当前已采用的竞赛状态集合 `draft / registration / running / frozen / ended`，以及 `manual_update / time_window` 等迁移原因和副作用状态 `pending / succeeded / failed`
  - 不负责：承诺额外状态枚举已经落地，或让页面自行扩展状态含义脱离后端模型

- `code/backend/internal/module/contest/infrastructure/contest_status_update_repository.go`、`contest_status_transition_repository.go`
  - 负责：在事务内更新 `contests.status`，并持久化 `contest_status_transitions` 作为迁移审计与副作用重放依据
  - 不负责：允许后台命令或脚本直接改 `contests` 行而绕开 transition 记录

- `code/backend/internal/module/contest/application/jobs/status_update_runner.go`、`lock_keepalive.go`、`status_transition_service.go`
  - 负责：按时间窗自动推进状态、持有并续租调度锁、跳过已被其他实例持有的 scheduler，并为 side effect replay 提供稳定入口
  - 不负责：依赖外部 cron 在数据库外维护一套平行状态机，或让锁过期后出现多实例并发推进

- `code/backend/internal/module/contest/application/commands/contest_update_commands.go`、`scoreboard_admin_freeze_commands.go`
  - 负责：承接管理员手动修改或冻结操作，并复用同一套 transition 协议和副作用 runner
  - 不负责：绕过状态机直接写榜单缓存、实例清理或公告广播逻辑

## 接口或数据影响

- 当前状态机直接影响 `contests.status`、`contest_status_transitions.side_effect_status`、榜单冻结逻辑以及 AWD 自动开赛 / 轮次调度窗口。
- 管理端更新赛事、冻结榜单和只读查询仍通过 `docs/contracts/openapi-v1.yaml` 中的 contest / scoreboard 契约暴露；接口只消费状态机结果，不重定义状态枚举。
- Redis 中只保留调度锁和辅助缓存，锁 key 见 `code/backend/internal/pkg/redis/keys.go`，不会替代 PostgreSQL 中的 transition audit。

## Guardrail

- migration 与状态迁移表：`code/backend/internal/app/contest_status_transition_migration_test.go`
- 时间口径与 contest 时间字段：`code/backend/internal/app/contest_time_migration_test.go`
- 自动调度与锁行为：`code/backend/internal/module/contest/application/jobs/status_updater_test.go`
- 手动状态更新共享同一协议：`code/backend/internal/module/contest/application/commands/contest_service_test.go`

## 历史迁移

- 这篇文档的 adopted 事实已经收口到五态状态机和 `contest_status_transitions` 审计表。
- 下文保留的背景、目标和分阶段落地只作为历史说明；若与 `code/backend/internal/model/contest.go`、`status_update_runner.go` 冲突，以当前代码为准。

## 0. 落地状态（2026-05-03）

当前仓库已经落地以下内容：

- `StatusUpdater` 使用分布式锁续租，持锁执行超过 TTL 时不会自然失锁。
- 自动状态推进改为 `ContestStatusTransition` 条件迁移，提交条件包含 `id + status + status_version`。
- 自动调度只在 `Applied=true` 后执行封榜快照和 AWD 运行态清理。
- `contests.status_version` 已用于标识每次成功状态迁移的顺序版本。
- `contest_status_transitions` 已持久化成功迁移和副作用结果。
- 人工管理入口在状态变化时也会递增 `status_version` 并记录迁移。

当前仍明确延后的内容：

- 没有新增完整 outbox 链路。当前副作用仍然是 Redis 本地幂等操作，等出现外部通知、异步审计或跨服务投递需求后再补 outbox。

## 1. 背景

当前竞赛状态由 `StatusUpdater` 定时扫描推进，状态包括：

- `draft`
- `registration`
- `running`
- `frozen`
- `ended`

现有代码已经在 `contest/domain/contest.go` 中定义了允许迁移关系，但定时任务的落库方式仍然更接近“扫描后直接写状态”：

- 查询满足时间窗口的比赛
- 在内存中计算目标状态
- 调用 `UpdateStatus(id, status)` 写入新状态
- 根据状态变化执行冻结快照、AWD 运行态清理等副作用

这套流程在单实例、低并发下能工作，但它没有把“状态迁移成功”和“副作用执行”绑定成一个严格的幂等单元。多 `ctf-api` 实例部署时，分布式锁可以减少重复扫描和重复推进，但正确性不能只依赖锁。

本设计的目标是把竞赛状态推进收敛成严格幂等的状态机：即使锁失效、任务重跑、多个实例同时尝试推进，同一场比赛的同一次状态迁移也只会被成功消费一次。

## 2. 设计目标

1. 状态迁移必须显式校验来源状态和目标状态。
2. 旧任务、重复任务或并发任务不能把比赛状态回退或重复消费同一迁移。
3. 状态迁移成功后才能执行对应副作用。
4. 副作用必须可重试，重复执行不会破坏结果。
5. 分布式锁仍然保留，但只承担多实例调度排他、降噪和减负载，不作为唯一正确性边界。
6. 当前 API 状态枚举不变，避免对前端和外部调用方产生兼容性影响。

## 3. 状态与允许迁移

状态机的合法迁移沿用现有领域规则：

```text
draft -> registration
registration -> draft
registration -> running
running -> frozen
running -> ended
frozen -> running
frozen -> ended
ended -> terminal
```

自动调度器只负责时间驱动迁移：

```text
registration -> running
running -> frozen
running -> ended
frozen -> ended
```

人工管理操作可以保留更宽的业务入口，例如封榜回退到 `running`，但必须复用同一套迁移校验和迁移记录，不允许绕过状态机直接写字符串状态。

## 4. 核心不变量

1. `ended` 是终态，任何自动任务都不能从 `ended` 迁出。
2. 自动调度不得把状态回退，例如不能把 `running` 写回 `registration`。
3. 每次迁移必须声明 `from_status`、`to_status` 和 `from_status_version`，提交时用数据库条件更新校验当前状态和版本仍与读取时一致。
4. 只有条件更新影响 1 行时，才认为本实例拥有这次迁移的副作用执行权。
5. 条件更新影响 0 行不等于失败，通常表示迁移已经被其他实例消费，当前任务应跳过副作用。

建议的写入语义：

```sql
UPDATE contests
SET status = $to_status,
    status_version = status_version + 1,
    updated_at = $now
WHERE id = $contest_id
  AND status = $from_status
  AND status_version = $from_status_version
  AND deleted_at IS NULL;
```

当前实现已经使用 `status_version` 建立 compare-and-set 语义，而不是只比较旧状态。

## 5. 迁移执行模型

状态推进应拆成三个明确步骤：

1. 计算候选迁移。
2. 提交条件迁移。
3. 仅在迁移提交成功后执行副作用。

推荐的应用层形态：

```go
transition := ContestStatusTransition{
    ContestID:  contest.ID,
    FromStatus: contest.Status,
    ToStatus:   nextStatus,
    Reason:     "time_window",
    OccurredAt: now,
}

result, err := transitionService.Apply(ctx, transition)
if err != nil {
    return err
}
if !result.Applied {
    return nil
}

sideEffects.Dispatch(ctx, result)
```

`Apply` 的职责是完成状态校验和条件写入。副作用不能在 `Apply` 之前执行，也不能只根据内存中的 `newStatus != contest.Status` 执行。

## 6. 副作用绑定

当前状态迁移关联的副作用主要有两类：

- `running -> frozen`：创建封榜快照
- `frozen -> running`：删除封榜快照
- `running -> ended`、`frozen -> ended`：清理比赛结束后的 AWD 运行态缓存

副作用执行原则：

1. 只有成功提交迁移的实例可以触发副作用。
2. 副作用处理必须幂等。
3. 副作用失败不能回滚已经提交的状态迁移，但必须可观察、可重试。

`running -> frozen` 的封榜快照建议使用确定性写入：

- Redis 快照 key 使用 `contest_id` 作为稳定 key。
- 快照内容由当前排行榜重新生成，重复生成结果应一致。
- 如需防止结束后被晚到任务覆盖，可补充 `frozen_snapshot_generated_at` 或迁移记录约束。

`ended` 清理建议保持幂等：

- Redis `DEL` 本身可重复执行。
- 清理失败记录错误日志，并由下一轮调度或专门修复任务重试。
- 不应因为清理失败把比赛状态回滚到非结束态。

如果后续副作用扩展到消息、审计、通知或外部系统，应引入 outbox：

- 状态迁移和 outbox 事件在同一数据库事务中提交。
- 后台 publisher 负责投递事件。
- consumer 端按 `transition_id` 去重。

## 7. 分布式锁定位

多 `ctf-api` 实例部署时仍然需要 `contest:status_updater:lock`。

锁的职责是：

- 避免多个实例同时扫描同一批候选比赛。
- 降低数据库和 Redis 写放大。
- 减少重复日志和无意义的条件更新尝试。
- 让定时任务运行行为更容易观测。

锁不承担唯一正确性保证。正确性由状态机的条件迁移和副作用幂等保证。

因此锁需要满足：

- 获取锁失败时跳过本轮扫描。
- 持锁执行超过 TTL 时自动续租。
- 续租发现 token 不匹配时取消当前运行上下文。
- 释放锁使用短超时上下文，避免上层取消导致释放失败。

这和 AWD round scheduler 的锁续租模型一致，但状态机本身必须能承受锁失效或重复执行。

## 8. 数据模型建议

当前实现已经包含两部分：

- `contests.status_version`：记录比赛已成功发生过多少次状态迁移。
- `contest_status_transitions`：记录每次成功迁移及其副作用结果。

迁移记录表结构：

```text
contest_status_transitions
- id
- contest_id
- status_version
- from_status
- to_status
- reason
- occurred_at
- applied_by
- side_effect_status
- side_effect_error
- created_at
```

当前唯一约束：

```text
(contest_id, status_version)
```

这里不能使用 `(contest_id, from_status, to_status)`，因为人工回退后再次发布会合法重复同一状态对。

## 9. 调度器流程

目标流程：

```text
StatusUpdater.Start
  -> acquire distributed lock
  -> start lock keepalive
  -> list transition candidates
  -> for each contest:
       calculate next status
       validate transition
       apply compare-and-set transition
       if applied:
         run idempotent side effects
       else:
         skip
  -> release lock
```

候选查询仍可按当前方式保留：

- `registration` 且 `start_time <= now`
- `running` 且 `freeze_time <= now` 或 `end_time <= now`
- `frozen` 且 `end_time <= now`

但候选查询只用于减少扫描范围，不能替代提交时的条件校验。

## 10. 测试要求

实现状态机时至少补充以下测试：

1. 并发两次 `registration -> running`，只能有一次返回 `applied=true`。
2. 第二个并发调用返回 stale/skip 时，不执行副作用。
3. `running -> frozen` 成功后只生成一次封榜快照。
4. `running -> ended` 和 `frozen -> ended` 的 AWD 缓存清理可重复执行。
5. `ended` 不能被自动任务迁出。
6. 非法迁移返回明确错误，例如 `registration -> ended` 不能绕过 `running`。
7. 分布式锁续租失败时，本轮调度停止继续处理后续比赛。

## 11. 分阶段落地

第一阶段：

- 保留现有分布式锁。
- 给 `contest_status_updater` 增加锁续租。
- 将仓储更新改为 `WHERE id = ? AND status = ?` 条件迁移。
- 仅在 `applied=true` 时执行副作用。

第二阶段：

- 引入 `ContestStatusTransitionService`，收口自动调度和人工管理入口。
- 明确返回 `applied / stale / invalid / not_found`。
- 把状态迁移校验从零散调用收敛到统一入口。

第三阶段后续扩展：

- 如果副作用扩展到外部系统，引入 outbox。
- 为状态迁移补充可观测指标，例如 `transition_applied_total`、`transition_stale_total`、`transition_side_effect_failed_total`。

## 12. 设计结论

竞赛状态推进应采用“双层保护”：

- 状态机保证正确性：条件迁移、合法状态校验、副作用幂等。
- 分布式锁保证运行效率：多实例排他、减少重复扫描和写放大。

锁失效时系统最多变重、变吵，不能变错。状态机重复执行时系统最多跳过 stale 迁移，不能重复消费副作用。
