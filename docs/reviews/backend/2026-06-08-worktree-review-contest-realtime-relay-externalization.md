# Review 对象

- 仓库：`/home/azhi/workspace/projects/.worktrees/ctf/2026-06-07-contest-realtime-relay-externalization`
- 分支：`task/2026-06-07-contest-realtime-relay-externalization`
- Task slug：`2026-06-07-contest-realtime-relay-externalization`
- Plan：`docs/plan/impl-plan/2026-06-07-contest-realtime-relay-externalization-implementation-plan.md`
- diff 来源：`main...HEAD`
- 评审范围：
  - `code/backend/internal/module/contest/**`
  - `code/backend/internal/module/ops/**`
  - `code/backend/internal/app/composition/ops_module.go`
  - `code/backend/internal/app/router_admin_contest_participation_routes.go`
  - `code/backend/internal/app/router_user_contest_routes.go`
  - `code/backend/migrations/000013_create_contest_realtime_outbox.*`
  - `code/frontend/src/api/contest*.ts`
  - `code/frontend/src/features/contest-announcements/**`
  - `code/frontend/src/features/contest-detail/**`
  - `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
  - `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
  - `feedback/2026-06-07-backoff-wait-must-not-depend-on-logger-presence.md`
  - `feedback/2026-06-07-contest-realtime-outbox-worker-must-not-revert-sent-rows-on-stale-retries.md`

# 分类检查

- 结论：同意 `非琐碎任务` 分类。
- 原因：
  - 同时触达 `contest -> ops -> app/composition -> frontend` 多个 owner 面。
  - 引入了 PostgreSQL outbox、Redis Streams、background dispatcher / consumer、HTTP 查询补读和前端公告/详情接线。
  - 这类变更即使测试通过，也不能只靠实现上下文自检直接宣布可合并。

# Gate Verdict

- 结论：`blocked`
- 原因：
  - 本轮 same-context self-check 没有发现新的 material code finding。
  - 但这不是独立 reviewer；根据 `code-workflow` 协议，独立 review gate 仍然未满足，当前不能把这份 self-check 当成最终 gate。

# Findings

## Blocker

1. 独立 review gate 尚未满足。
   - 依据：
     - `~/.agents/harness/workflows/code-workflow/independent-review-protocol.md` 明确要求 `非琐碎任务` 在 `completion-full` 后进入独立 `code-reviewer` review。
     - 当前这份记录来自实现上下文本地自检，不是独立 reviewer。
   - 影响：
     - 目前只能说明“实现上下文下的代码和验证证据看起来一致”，还不能作为最终 merge gate。
   - 收口方向：
     - 交给独立 reviewer 按同一份 plan、commit range 和验证证据做 gate review。

# Material Findings

- 代码层面：本轮 self-check 未发现新的 material finding。
- 流程层面：独立 review gate 未满足，仍是当前唯一 blocker。

# Non-blocking Suggestions

- 当前没有额外需要阻塞的次要建议。

# 已复核的关键点

- `contest` 继续拥有领域事件和 outbox 记账，`ops` 继续拥有 relay adapter / stream publish / consume owner，分层方向与 plan 一致。
- `AWD preview progress` 已不再依赖旧的即时 event bus 接线，测试已对齐到当前 outbox owner。
- `main` 回灌时暴露的 `awd_service_test.go` 冲突已按主线现状收口到拆分后的 `awd_service_preview_test.go` / `awd_service_attack_test.go` 结构，没有把旧大文件重新并回。

# Required Re-validation

本轮 self-check 已实际执行：

```bash
cd code/backend && go test ./internal/module/contest/infrastructure -run 'TestRealtimeOutboxRepository' -count=1
cd code/backend && go test ./internal/module/ops/application/commands -run 'TestContestRealtime(Service|OutboxDispatcher)' -count=1
cd code/backend && go test ./internal/module/ops/infrastructure -run 'TestContestRealtimeStream' -count=1
cd code/backend && go test ./internal/module/contest/application/commands -count=1
cd code/backend && go test ./internal/module/contest/application/queries -count=1
cd code/backend && go test ./internal/app -run 'Test(BuildOpsModuleDelegatesToContainerRuntime|BuildContestModuleDelegatesToRuntime)' -count=1
cd code/frontend && npx vitest run \
  src/features/contest-announcements/model/useAnnouncementSubscription.test.ts \
  src/features/contest-announcements/model/useContestAnnouncementManagement.test.ts \
  src/pages/contests/__tests__/ContestDetail.awd.test.ts \
  src/pages/contests/__tests__/ContestDetail.challenge-flow.test.ts
```

独立 reviewer 至少应复用或抽查上述证据，并按风险决定是否还要补跑：

```bash
bash scripts/run-workflow-stage.sh completion-full
```

# Senior implementation assessment

- 当前实现方向是成立的：没有把分布式 relay 语义硬塞回 `platform/events.Bus`，而是把 durable relay 账本留在 `contest`，把跨实例转发落到 `ops`。
- 相比继续依赖“业务成功后顺手 best-effort 广播”，这版结构更接近仓库里已有的“事务内记账 + 后台恢复副作用”模式，owner 更清楚，失败面也更可测。
- 前端侧保持“轻事件通知 + HTTP 刷新/补查”而没有顺手扩成客户端 cursor replay，也符合这次任务的第一刀边界。

# Residual Risk

- 本轮没有真实执行多实例 runtime 场景下的手工恢复演练；实例级恢复主要由 repository / stream / service tests 间接证明。
- `completion-full` 当前只对“当前未提交改动”敏感，不会自动覆盖 `main...HEAD` 的整个已提交差异；因此这次 merge 准备主要依赖手工挑选的包级验证。
- 项目 skill 中提到的 `references/ctf-current-review-status-checks.md`、`references/technical-risk-checks.md`、`references/test-strategy-review.md` 在仓库里不存在，本轮无法把它们作为事实源引用。

# Touched known-debt status

- 本轮触达的是 `contest realtime relay` 自身的 owner 边界，没有发现“已登记且本次 touched surface 必须顺手关闭、但仍未关闭”的额外结构债。
- `AWD preview progress` 测试 owner 漂移是本次回灌后暴露出来的 touched-surface 问题，已在当前分支内收口，没有留成 follow-up。
