# Reuse Decision

## Change type
structure_refactor

## Existing code searched
- `code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue`
- `code/frontend/src/components/challenge/ChallengeActionAside.vue`
- `code/frontend/src/components/challenge/ChallengeQuestionPanel.vue`
- `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
- `code/frontend/src/components/challenge/ChallengeInstanceCard.vue`
- `code/frontend/src/features/challenge-list/model/index.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeWorkspaceShell.vue`
- `code/frontend/src/features/challenge-detail/ui/index.ts`
- `code/frontend/src/pages/challenges/ChallengeListRoutePage.vue`
- `code/frontend/src/pages/challenges/__tests__/ChallengeList.test.ts`
- `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`
- `code/frontend/src/pages/__tests__/sharedPaginationControls.test.ts`
- `code/frontend/src/pages/__tests__/journalUserDirectoryStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-28-challenge-detail-feature-ui-batch-review.md`

## Similar implementations found
- `code/frontend/src/features/platform/challenge-detail/ui/AdminChallengeTopbarPanel.vue`
- `code/frontend/src/features/platform/challenge-detail/ui/AdminChallengeWorkspaceTabs.vue`
- `code/frontend/src/features/platform/challenge-detail/ui/AdminChallengeProfilePanel.vue`
- `code/frontend/src/features/platform/challenges/ui/ChallengeManageDirectoryPanel.vue`
- `code/frontend/src/features/contest-detail/ui/ContestOverviewPanel.vue`

## Decision
refactor_existing

## Reason
这次不是新增 UI，而是把 owner 已经明确的 challenge capability UI 从历史 `components/challenge` 目录回收到对应 feature。`ChallengeDirectoryPanel.vue` 现在只被 `ChallengeListRoutePage.vue` 通过 `challenge-list` page model 驱动，适合直接落到 `features/challenge-list/ui`。`ChallengeActionAside.vue`、`ChallengeQuestionPanel.vue`、`ChallengeWriteupPanel.vue` 和 `ChallengeInstanceCard.vue` 都只服务 `features/challenge-detail/ui/ChallengeWorkspaceShell.vue` 这条 student challenge detail workspace 装配链，继续留在旧组件目录只会让 raw-source 测试、自动组件声明和 backlog 事实继续指向错误 owner。最小正确改动是复用现有实现与 props / emits contract，仅迁移文件落点、feature public API、测试护栏和 backlog 记录，不重做交互和样式结构。

## Files to modify
- `.harness/reuse-decisions/challenge-feature-boundary-cleanup.md`
- `docs/plan/impl-plan/2026-05-30-challenge-feature-boundary-cleanup-plan.md`
- `docs/reviews/frontend/2026-05-30-challenge-feature-boundary-cleanup-review.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/pages/challenges/ChallengeListRoutePage.vue`
- `code/frontend/src/pages/challenges/__tests__/ChallengeList.test.ts`
- `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`
- `code/frontend/src/pages/__tests__/sharedPaginationControls.test.ts`
- `code/frontend/src/pages/__tests__/journalUserDirectoryStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/features/challenge-list/index.ts`
- `code/frontend/src/features/challenge-list/ui/index.ts`
- `code/frontend/src/features/challenge-list/ui/ChallengeDirectoryPanel.vue`
- `code/frontend/src/features/challenge-detail/ui/index.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeWorkspaceShell.vue`
- `code/frontend/src/features/challenge-detail/ui/ChallengeActionAside.vue`
- `code/frontend/src/features/challenge-detail/ui/ChallengeQuestionPanel.vue`
- `code/frontend/src/features/challenge-detail/ui/ChallengeWriteupPanel.vue`
- `code/frontend/src/features/challenge-detail/ui/ChallengeInstanceCard.vue`
- `code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue`
- `code/frontend/src/components/challenge/ChallengeActionAside.vue`
- `code/frontend/src/components/challenge/ChallengeQuestionPanel.vue`
- `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
- `code/frontend/src/components/challenge/ChallengeInstanceCard.vue`

## After implementation
- 如果后续还有 challenge 私有展示块继续留在 `components/challenge/*`，优先按 `challenge-list` / `challenge-detail` / `entities/challenge` 三个 owner 重新判断，而不是再新增历史目录例外。
