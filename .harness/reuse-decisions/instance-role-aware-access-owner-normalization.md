# Reuse Decision

## Change type
frontend architecture / role-aware access owner normalization

## Existing code searched
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/awd-reviews.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`

## Similar implementations found
- `api/awd-reviews.ts` 已经作为 AWD review 共享 feature 的中立 role-aware access owner，统一承接 admin / teacher API 选择。
- 实例目录 contract 命名上一刀已收口到 `InstanceDirectory*`，当前残余 teacher 语义主要只剩 public wrapper 与 feature access owner。
- `api/admin/teaching.ts` 和 `api/teacher/instances.ts` 当前都只是薄 wrapper，说明 role-aware access owner 还有进一步集中空间。

## Decision
refactor_existing

## Reason
- `usePlatformInstanceManagementPage.ts` 与 `useInstances.ts` 现在分别直连 `@/api/admin` 和 `@/api/teacher` 的实例目录函数，shared domain workflow 仍然要感知角色历史 owner。
- 最小正确改动是新增一层中立 `api/instances.ts`，统一承接实例目录 list / destroy 的 role-aware API 选择，再让 teacher / platform 实例目录 feature 都只依赖这层 facade。
- 本轮不继续改 `getTeacherInstances()`、`destroyTeacherInstance()`、`getPlatformInstances()` 的 public wrapper 名，也不改后端 `/api/v1/teacher/instances` path，blast radius 可控。

## Files to modify
- `.harness/reuse-decisions/instance-role-aware-access-owner-normalization.md`
- `docs/plan/impl-plan/2026-05-28-instance-role-aware-access-owner-normalization-plan.md`
- `docs/reviews/frontend/2026-05-28-instance-role-aware-access-owner-normalization-review.md`
- `code/frontend/src/api/instances.ts`
- `code/frontend/src/api/__tests__/instances.test.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 实例目录共享 workflow 会把 role-aware API 选择收口成单点 owner。
- teacher / platform 实例目录 feature 不再各自直连 `@/api/admin` / `@/api/teacher` 的实例目录函数。
- admin / teacher 结构耦合在实例 access owner 这一层会继续向 transport/public wrapper 收缩。
