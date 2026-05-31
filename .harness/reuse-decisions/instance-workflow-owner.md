# Reuse Decision

## Change type
frontend refactor / instance workflow owner convergence

## Existing code searched
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `code/frontend/src/features/instance-list/model/useInstanceOperations.ts`
- `code/frontend/src/features/instance-list/model/useInstanceListPage.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAccessActions.ts`
- `code/frontend/src/pages/instances/__tests__/InstanceList.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`

## Similar implementations found
- `challenge-detail` 已经本地持有 student instance 的 `open / extend / destroy` workflow，并附带创建后的 polling owner。
- `instance-list` 已经本地持有 student instance 的 `open / extend / destroy` workflow，并附带列表回填、warning dialog 联动 owner。
- `contest-awd-workspace` 也有 `openService / openTarget` 访问动作，但它是 contest-aware access workflow，不适合作为普通实例 workflow 的 owner。

## Decision
refactor_existing

## Reason
当前最小正确切片不是继续在 `challenge-detail` 和 `instance-list` 各自修补，而是把普通实例动作 owner 收口成一个共享 feature：

- 两条链路都在调用 `requestInstanceAccess / extendInstance / destroyInstance`
- 都有相同的 TCP 复制逻辑、统一危险确认、销毁错误消息优先级
- 差异主要只剩页面本地的目标解析和成功后的状态回填

因此本轮应：

- 新建 `features/instance-workflow`
- 让它承接普通实例的 `open / extend / destroy` 动作流程
- `challenge-detail` 只保留 `create / refresh / polling` owner
- `instance-list` 只保留列表 mutation、warning state 和 manual-action policy

本轮不做：

- 不把 workflow 抽进 `entities/instance`
- 不处理 teacher / platform 的 managed instance destroy owner
- 不改 AWD workspace 的 contest-aware access action

## Files to modify
- `.harness/reuse-decisions/instance-workflow-owner.md`
- `docs/plan/impl-plan/2026-05-31-instance-workflow-owner-plan.md`
- `code/frontend/src/features/instance-workflow/index.ts`
- `code/frontend/src/features/instance-workflow/model/index.ts`
- `code/frontend/src/features/instance-workflow/model/useInstanceWorkflowActions.ts`
- `code/frontend/src/features/instance-workflow/model/useInstanceWorkflowActions.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.ts`
- `code/frontend/src/features/instance-list/model/useInstanceOperations.ts`

## After implementation
- 普通实例的共享动作 owner 会集中到 `features/instance-workflow`
- `challenge-detail` 和 `instance-list` 不再各自维护重复的访问、延时和销毁流程
- 后续 teacher / platform 若要继续迁移，可以直接决定是否复用同一 workflow feature，或者另建 managed-instance workflow
