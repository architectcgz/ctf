# Utility Owner Realignment 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-utility-owner-realignment-plan.md`
- files reviewed：
  - `.harness/reuse-decisions/utility-owner-realignment.md`
  - `docs/plan/impl-plan/2026-05-29-utility-owner-realignment-plan.md`
  - `docs/reviews/frontend/2026-05-29-utility-owner-realignment-review.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/entities/contest/**`
  - `code/frontend/src/entities/skill-profile/**`
  - `code/frontend/src/entities/contest-awd-challenge-link/**`
  - `code/frontend/src/api/assessment.ts`
  - `code/frontend/src/api/teacher/students.ts`
  - `code/frontend/src/api/teaching/students.ts`
  - 相关 consumer import 变更
- Classification check：同意按 owner realignment 处理；风险主要在旧 util 删除后是否还存在漏改 import，以及新 entity owner 是否足够清晰。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `contest` 相关展示语义已经回到 `entities/contest`，不再伪装成无归属的 `utils/contest.ts`。
- `skill profile / recommendation` 的 normalize 与展示辅助已经回到 `entities/skill-profile`，API wrapper 与 feature consumer 继续走稳定 public API。
- AWD service 到 challenge link 的共享 mapper 已迁到 `entities/contest-awd-challenge-link`，比挂在任一单独 feature 下更贴合 cross-feature owner。
- 旧 `utils/contest.ts`、`utils/skillProfile.ts`、`utils/platformContestAwdChallengeLinks.ts` 已退出主路径，没有保留 re-export bridge。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/contests/__tests__/ContestDetail.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/utility-owner-realignment.md docs/plan/impl-plan/2026-05-29-utility-owner-realignment-plan.md docs/reviews/frontend/2026-05-29-utility-owner-realignment-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/entities/contest code/frontend/src/entities/skill-profile code/frontend/src/entities/contest-awd-challenge-link code/frontend/src/api/assessment.ts code/frontend/src/api/teacher/students.ts code/frontend/src/api/teaching/students.ts code/frontend/src/components/contests/ContestOverviewPanel.vue code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue code/frontend/src/features/contest-awd-admin/model/useAwdContestSnapshotLoader.ts code/frontend/src/features/contest-awd-admin/model/useAwdChallengeLinkOperations.ts code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts code/frontend/src/features/contest-detail/model/useContestListPage.ts code/frontend/src/features/contest-workbench/model/useContestChallengeOrchestration.ts code/frontend/src/features/contest-workbench/model/useContestEditAwdWorkspace.ts code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue code/frontend/src/features/platform-contests/ui/PlatformContestRulesSection.vue code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts code/frontend/src/features/scoreboard/model/useScoreboardContestDirectoryPage.ts code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisDataState.ts code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts code/frontend/src/features/teacher-workspace/model/useWorkspace.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/utils/contest.ts code/frontend/src/utils/skillProfile.ts code/frontend/src/utils/platformContestAwdChallengeLinks.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；独立 reviewer gate 仍未满足。
- `featureRouterImportAllowlist` 与 `composableMultiBoundaryAllowlist` 仍然保留，属于下一轮单独判定范围，不和这轮 owner realignment 混处理。
