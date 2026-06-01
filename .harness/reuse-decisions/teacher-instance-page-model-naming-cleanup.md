# Reuse Decision

## Change type
frontend refactor / teacher instance page model naming cleanup

## Existing code searched
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/pages/teacher/InstanceManagementRoutePage.vue`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- 平台实例页 page model 已命名为 `usePlatformInstanceManagementPage()`，route page 和 raw-source 护栏都直接体现 owner。
- 教师实例 feature 当前同时存在 `useInstances()` 和 `useInstanceManagementPage()`，前者命名过泛，后者又没有体现 teacher owner，和当前 teacher-specific feature 边界不一致。
- 最近几轮 student/class/dashboard/instance 迁移都在通过命名把 page owner、route owner 和共享 workflow owner 区分开。

## Decision
refactor_existing

## Reason
下一轮最小正确切片不是继续拆教师实例 workflow，而是先把 teacher feature 内部 page-state owner 命名收紧：

- `useInstances()` 改成显式 teacher owner 命名，避免和 student instance list、challenge instance workflow、managed instance directory 混淆。
- `useInstanceManagementPage()` 也改成带 teacher owner 的 page-model 命名，让 route page 和 raw-source 护栏直接体现边界。
- 不移动文件路径，不改 public route、filter、destroy、pagination 或 shared managed-instance workflow 行为。

## Files to modify
- `.harness/reuse-decisions/teacher-instance-page-model-naming-cleanup.md`
- `docs/plan/impl-plan/2026-06-01-teacher-instance-page-model-naming-cleanup-plan.md`
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/pages/teacher/InstanceManagementRoutePage.vue`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- 教师实例 route page 会直接暴露 teacher-specific page model owner 命名。
- teacher feature 内部不再把 page-state owner 继续挂在过泛的 `useInstances()` 名下。
