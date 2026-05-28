# Reuse Decision

## Change type
+api / feature / component / docs / test

## Existing code searched
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- 共享班级目录 DTO 已按最小切片从 `TeacherClassItem` 收口到 `ClassDirectoryItem`，说明当前可以继续沿“共享 contract 一组一刀”的方式推进。
- AWD review 共享 DTO 也已经按 `AwdReview*` 中性命名完成收口，teacher / platform 只在 public wrapper 和 HTTP path 层保留角色语义。
- 实例目录链路当前同时被 teacher 和 platform feature 消费，`TeacherInstance*` 前缀已经不再符合真实 owner 语义。

## Decision
refactor_existing

## Reason
- `TeacherInstanceItem`、`TeacherInstancePageData`、`TeacherInstanceListSummaryData` 与 `TeacherInstanceStatusFilter` 已经是跨角色共享 contract，继续保留 teacher 前缀会放大 admin / teacher 结构耦合的残余语义。
- 这次只收“实例目录 contract 命名”本身，不顺手改 `getTeacherInstances()`、`destroyTeacherInstance()` 这类 public wrapper owner，也不混入其它 student / overview DTO，blast radius 可控。

## Files to modify
- `.harness/reuse-decisions/instance-directory-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-28-instance-directory-contract-naming-neutralization-implementation-plan.md`
- `docs/reviews/frontend/2026-05-28-instance-directory-contract-naming-neutralization-review.md`
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-instances/model/useInstances.ts`
- `code/frontend/src/features/teacher-instances/ui/TeacherInstanceManagementPage.vue`
- `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 这组实例目录 contract 收口后，前端本地更深层的 teacher 前缀共享 contract 将进一步收敛到更少数的历史 DTO，而不是继续停留在 teacher / platform 都会直接消费的目录流上。
