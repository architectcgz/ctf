# Reuse Decision

## Change type
frontend refactor / challenge detail workspace tab owner convergence

## Existing code searched
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
- `code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue`
- `code/frontend/src/shared/model/navigation/useRouteQueryTabs.ts`
- `code/frontend/src/shared/model/navigation/useUrlSyncedTabs.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `student-analysis`、`class-students`、`teacher-dashboard` 已经把 `panel` query owner 收口到 page model，并让 UI / route shell 退回展示和键盘焦点契约。
- `contest-detail` 已经在 route page 里从 `useUrlSyncedTabs()` 切到 `useRouteQueryTabs()`，说明 contest / challenge 这类 query-tab owner 可以直接复用共享 helper。
- `useChallengeDetailPage.ts` 当前仍直接使用 `useUrlSyncedTabs()`，并把 workspace tab 的 button ref / keydown 一起留在 page model。

## Decision
refactor_existing

## Reason
下一轮最小正确切片不是改 challenge detail 的加载或题解 workflow，而是把 workspace tab owner 对齐到当前共享 query-tab 模式：

- `useChallengeDetailPage.ts` 改为复用 `useRouteQueryTabs()` 承接 `question/solution/records/writeup` 的 `panel` query。
- `ChallengeDetailRoutePage.vue` / `ChallengeDetailWorkspace.vue` 只保留展示桥接和必要的键盘导航契约，不再让 page model 继续绑定 `useUrlSyncedTabs()` 那套本地 query-tab helper。
- 不改 solution tab、实例 workflow、题解 / 提交记录 / writeup 加载策略。

本轮不做：

- 不改 `useChallengeDetailInteractions.ts`、`useChallengeDetailDataLoader.ts`、`useChallengeInstance.ts` 的异步 owner。
- 不改 platform challenge detail 那条已有 `useRouteQueryTabs()` 的实现。
- 不扩到 `contest-edit` 的 stage owner。

## Files to modify
- `.harness/reuse-decisions/challenge-detail-workspace-tab-owner-convergence.md`
- `docs/plan/impl-plan/2026-05-31-challenge-detail-workspace-tab-owner-convergence-plan.md`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
- `code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue`
- `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `useChallengeDetailPage.ts` 会改成 challenge detail workspace `panel` query 的唯一 owner。
- challenge detail user route / widget 会对齐到 `page model + shared route query owner + 展示壳 / 键盘契约` 这套模式。
