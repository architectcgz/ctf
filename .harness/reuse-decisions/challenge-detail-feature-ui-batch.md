# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- `code/frontend/src/components/challenge/ChallengeSolutionsPanel.vue`
- `code/frontend/src/components/challenge/ChallengeSubmissionRecordsPanel.vue`
- `code/frontend/src/views/challenges/ChallengeDetail.vue`
- `code/frontend/src/features/challenge-detail/index.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailSharedShell.test.ts`
- `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `platform-challenge-detail` 已证明，单一 feature 的 detail workspace 子面板应直接落到 `features/*/ui`，由 feature public API 暴露给 route view 或 widget。
- `platform-contests`、`platform-user-management` 与 `teacher-dashboard` 最近几轮迁移说明，这类 page shell / panel 迁位应同时更新 raw-source 护栏、allowlist 和 `components.d.ts`，不保留 legacy wrapper。
- `ChallengeDetail.vue` 当前已经把 route/query、加载与主动作 owner 留在 page，本轮迁位只需要继续把 feature 私有 UI 从旧 `components/challenge/*` 收回 `features/challenge-detail/ui`。

## Decision
refactor_existing

## Reason
- `ChallengeWorkspaceShell.vue`、`ChallengeSolutionsPanel.vue`、`ChallengeSubmissionRecordsPanel.vue` 只服务 `challenge-detail` 这一条 feature，不是跨页面复用的 shared challenge component。
- 继续停留在 `components/challenge/*` 会让 `componentFeatureImportAllowlist` 长期保留 3 条明显的单 feature 例外。
- 最小正确改动是把这组三件套整体迁入 `features/challenge-detail/ui`，并让 `ChallengeDetail.vue` 与子组件装配统一经由 feature public API / feature 内部相对 import 完成。

## Files to modify
- `.harness/reuse-decisions/challenge-detail-feature-ui-batch.md`
- `docs/plan/impl-plan/2026-05-28-challenge-detail-feature-ui-batch-plan.md`
- `docs/reviews/frontend/2026-05-28-challenge-detail-feature-ui-batch-review.md`
- `code/frontend/src/features/challenge-detail/index.ts`
- `code/frontend/src/features/challenge-detail/ui/index.ts`
- `code/frontend/src/features/challenge-detail/ui/ChallengeWorkspaceShell.vue`
- `code/frontend/src/features/challenge-detail/ui/ChallengeSolutionsPanel.vue`
- `code/frontend/src/features/challenge-detail/ui/ChallengeSubmissionRecordsPanel.vue`
- `code/frontend/src/views/challenges/ChallengeDetail.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailSharedShell.test.ts`
- `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `challenge-detail` 这 3 条 `componentFeatureImportAllowlist` 应该清空。
- `ChallengeDetail.vue` 不再直连旧 `components/challenge/ChallengeWorkspaceShell.vue`。
- `ChallengeSolutionsPanel.vue` 与 `ChallengeSubmissionRecordsPanel.vue` 作为 feature 私有 UI 不再滞留在 legacy challenge 组件目录。
