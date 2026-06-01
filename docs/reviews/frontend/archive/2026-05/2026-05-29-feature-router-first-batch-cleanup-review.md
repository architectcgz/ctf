# Feature Router First Batch Cleanup 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-feature-router-first-batch-cleanup-plan.md`
- Scope：
  - `useScoreboardDetailPage.ts`
  - `useAwdChallengeLibraryPage.ts`
  - `useRegisterPage.ts`
  - `useChallengeImportManagePage.ts`
  - `useChallengeImportPreviewPage.ts`
- Classification check：同意按“第一批低复杂度 route target cleanup”批量处理；这五条都属于 route param、薄导航或 success redirect，不需要先进入更重的 query/workflow owner 重构。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- 这批不应再新增新的 router helper 杂项；应尽量复用现有 `AppRouteLink` / `AppRouteRedirect` 和显式 route target contract。
- `scoreboard detail` 与 `challenge import preview` 的 route param 下沉为 props 后，对应 view 边界会更接近当前仓库的 route shell 模式。
- `register` 与 `import preview` 的“成功后跳转”仍然是 workflow owner，但 transport 不应继续停在 feature model 里。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/auth/__tests__/RegisterView.test.ts src/views/platform/__tests__/ChallengeImportManage.test.ts src/views/platform/__tests__/ChallengeImportPreview.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/feature-router-first-batch-cleanup.md docs/plan/impl-plan/2026-05-29-feature-router-first-batch-cleanup-plan.md docs/reviews/frontend/2026-05-29-feature-router-first-batch-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/router/routes/authRoutes.ts code/frontend/src/router/routes/studentRoutes.ts code/frontend/src/router/routes/platformRoutes.ts code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts code/frontend/src/features/auth/model/useRegisterPage.ts code/frontend/src/features/challenge-package-import/model/index.ts code/frontend/src/features/challenge-package-import/model/useChallengeImportManagePage.ts code/frontend/src/features/challenge-package-import/model/useChallengeImportPreviewPage.ts code/frontend/src/views/scoreboard/ScoreboardDetail.vue code/frontend/src/views/platform/AWDChallengeLibrary.vue code/frontend/src/views/auth/RegisterView.vue code/frontend/src/views/platform/ChallengeImportManage.vue code/frontend/src/views/platform/ChallengeImportPreview.vue code/frontend/src/components/platform/challenge/ChallengeImportHeroPanel.vue code/frontend/src/components/platform/challenge/ChallengeImportQueuePanel.vue code/frontend/src/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue code/frontend/src/features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts code/frontend/src/views/auth/__tests__/RegisterView.test.ts code/frontend/src/views/platform/__tests__/ChallengeImportManage.test.ts code/frontend/src/views/platform/__tests__/ChallengeImportPreview.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 第二批和第三批 allowlist 仍未处理。
- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
