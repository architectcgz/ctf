# Challenge Feature Boundary Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`refactor/challenge-feature-boundary-cleanup`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-30-challenge-feature-boundary-cleanup-plan.md`
- Scope：
  - `features/challenge-list/ui/ChallengeDirectoryPanel.vue`
  - `features/challenge-detail/ui/ChallengeActionAside.vue`
  - `features/challenge-detail/ui/ChallengeQuestionPanel.vue`
  - `features/challenge-detail/ui/ChallengeWriteupPanel.vue`
  - `features/challenge-detail/ui/ChallengeInstanceCard.vue`
  - route pages / raw-source tests / `components.d.ts`
- Classification check：同意按“已明确 owner 的 challenge feature UI 迁位”单独切片；这条的主要风险是路径迁移回归，而不是交互逻辑重构。
- Gate verdict：Pass with known baseline failures

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Verification summary

- `bash scripts/check-task-intake.sh --reuse-decision challenge-feature-boundary-cleanup`
  - 结果：通过。
- `cd code/frontend && npm run test:run -- src/pages/challenges/__tests__/ChallengeList.test.ts src/pages/challenges/__tests__/ChallengeDetail.test.ts src/pages/challenges/__tests__/challengePageUiStrategy.test.ts src/pages/__tests__/sharedPaginationControls.test.ts src/pages/__tests__/journalUserDirectoryStyles.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/features/__tests__/featureBoundaries.test.ts`
  - 结果：`ChallengeList.test.ts`、`ChallengeDetail.test.ts`、`challengePageUiStrategy.test.ts`、`sharedPaginationControls.test.ts`、`journalUserDirectoryStyles.test.ts`、`workspacePageHeaderStyles.test.ts` 通过；`featureBoundaries.test.ts` 失败。
- `cd /home/azhi/workspace/projects/ctf/code/frontend && npm run test:run -- src/features/__tests__/featureBoundaries.test.ts`
  - 结果：主工作区同样失败，确认这是仓库既有基线，不是本次迁位新增。
- `cd code/frontend && npm run typecheck`
  - 结果：失败，但报错落在未触达的既有文件 `src/features/platform/overview/ui/PlatformOverviewHeroPanel.vue` 与 `src/features/teaching/student-analysis-workspace/ui/StudentReviewWorkspace.vue`。
- `git diff --check -- ...`
  - 结果：通过。

## Review focus

- `challenge-list` 与 `challenge-detail` 是否都只通过 feature owner 消费新路径。
- raw-source 护栏和自动组件声明是否已经完全从旧 `components/challenge/*` 切走。
- 旧路径删除后是否还存在悬挂 import 或测试引用。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/pages/challenges/__tests__/ChallengeList.test.ts src/pages/challenges/__tests__/ChallengeDetail.test.ts src/pages/challenges/__tests__/challengePageUiStrategy.test.ts src/pages/__tests__/sharedPaginationControls.test.ts src/pages/__tests__/journalUserDirectoryStyles.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/features/__tests__/featureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/.worktrees/ctf-challenge-feature-boundary-cleanup && git diff --check -- .harness/reuse-decisions/challenge-feature-boundary-cleanup.md docs/plan/impl-plan/2026-05-30-challenge-feature-boundary-cleanup-plan.md docs/reviews/frontend/2026-05-30-challenge-feature-boundary-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/components.d.ts code/frontend/src/pages/challenges/ChallengeListRoutePage.vue code/frontend/src/pages/challenges/__tests__/ChallengeList.test.ts code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts code/frontend/src/pages/__tests__/sharedPaginationControls.test.ts code/frontend/src/pages/__tests__/journalUserDirectoryStyles.test.ts code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/features/challenge-list/index.ts code/frontend/src/features/challenge-list/ui/index.ts code/frontend/src/features/challenge-list/ui/ChallengeDirectoryPanel.vue code/frontend/src/features/challenge-detail/ui/index.ts code/frontend/src/features/challenge-detail/ui/ChallengeWorkspaceShell.vue code/frontend/src/features/challenge-detail/ui/ChallengeActionAside.vue code/frontend/src/features/challenge-detail/ui/ChallengeQuestionPanel.vue code/frontend/src/features/challenge-detail/ui/ChallengeWriteupPanel.vue code/frontend/src/features/challenge-detail/ui/ChallengeInstanceCard.vue`

## Residual risk

- `featureBoundaries.test.ts` 与全局 `typecheck` 当前都不是绿色基线，本次只能确认 challenge 迁位相关测试通过，不能把仓库整体判成全绿。
- 这次 review 仍是同上下文归档，不等价于独立 reviewer gate。
