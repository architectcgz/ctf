# Reuse Decision

## Change type
frontend refactor / managed instance destroy workflow convergence

## Existing code searched
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/features/platform/challenges/model/usePlatformChallenges.ts`

## Similar implementations found
- teacher 实例管理在 `useInstanceManagementPage` 做危险确认，在 `useInstances` 做 role-aware destroy + reload + toast。
- platform 实例管理在 `usePlatformInstanceManagementPage` 同时做危险确认、role-aware destroy 和 reload。
- 其它平台列表里也有“确认后删除”的模式，但没有 role-aware managed instance 的共享 owner。

## Decision
refactor_existing

## Reason
当前最小正确切片是把 teacher / platform 共用的 managed instance 销毁动作收口为单独 workflow feature：

- 两边都调用 `destroyManagedInstanceByRole`
- 都需要统一 destructive confirm
- 都要处理 `destroyingId`、删除后翻页回退和 reload
- 差异主要只剩确认文案、成功提示和错误展示策略

因此本轮应：

- 新建 `features/managed-instance-workflow`
- 让它承接 role-aware managed instance destroy workflow
- teacher / platform 只保留各自列表状态、reload 策略和页面级文案策略

本轮不做：

- 不扩展到普通 instance `open / extend / destroy`
- 不改 teacher / platform 的 list loading、filter 或 route owner
- 不处理 AWD workspace、contest service 等其它销毁流

## Files to modify
- `.harness/reuse-decisions/managed-instance-destroy-workflow-owner.md`
- `docs/plan/impl-plan/2026-05-31-managed-instance-destroy-workflow-owner-plan.md`
- `code/frontend/src/features/managed-instance-workflow/index.ts`
- `code/frontend/src/features/managed-instance-workflow/model/index.ts`
- `code/frontend/src/features/managed-instance-workflow/model/useManagedInstanceDestroyAction.ts`
- `code/frontend/src/features/managed-instance-workflow/model/useManagedInstanceDestroyAction.test.ts`
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`

## After implementation
- teacher / platform 不再各自维护 role-aware managed instance destroy 流程。
- `managed-instance-workflow` 会成为危险确认、role-aware destroy 和重复点击保护的共享 owner。
- teacher / platform 页面只保留各自目录数据 owner 和文案差异。
