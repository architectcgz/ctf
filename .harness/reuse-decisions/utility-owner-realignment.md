# Reuse Decision

## Change type
frontend refactor / utility owner realignment

## Existing code searched
- code/frontend/src/utils/contest.ts
- code/frontend/src/utils/skillProfile.ts
- code/frontend/src/utils/platformContestAwdChallengeLinks.ts
- code/frontend/src/entities/challenge/*
- code/frontend/src/features/contest-awd-admin/model/*
- code/frontend/src/features/contest-detail/model/useContestListPage.ts
- code/frontend/src/features/contest-workbench/model/*
- code/frontend/src/features/scoreboard/model/useScoreboardContestDirectoryPage.ts
- code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts
- code/frontend/src/api/assessment.ts
- code/frontend/src/api/teacher/students.ts
- code/frontend/src/api/teaching/students.ts

## Similar implementations found
- `entities/challenge` 已经作为展示语义和最小 contract owner 落地，说明“纯领域展示 helper / normalize helper”应回到 entity，而不是停在 `utils/`。
- `platformContestAwdChallengeLinks.ts` 这种被两个 feature 共用的 mapper，更接近 cross-feature entity / shared domain owner，而不是无归属的 `utils/*`。

## Decision
refactor_existing

## Reason
上一刀已经把 `contest.ts`、`skillProfile.ts`、`platformContestAwdChallengeLinks.ts` 的 contract 从 `@/api/contracts` 脱钩，但路径仍停在 `utils/*`。这会继续留下两个结构问题：

- 领域展示规则伪装成“通用工具”
- cross-feature AWD mapper 没有明确 owner

最小正确改动是：

- 新建 `entities/contest`
- 新建 `entities/skill-profile`
- 新建 `entities/contest-awd-challenge-link`
- 全量改 consumer import
- 删除旧 `utils/*` 文件

## Files to modify
- .harness/reuse-decisions/utility-owner-realignment.md
- docs/plan/impl-plan/2026-05-29-utility-owner-realignment-plan.md
- docs/reviews/frontend/2026-05-29-utility-owner-realignment-review.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/entities/contest/**
- code/frontend/src/entities/skill-profile/**
- code/frontend/src/entities/contest-awd-challenge-link/**
- code/frontend/src/api/assessment.ts
- code/frontend/src/api/teacher/students.ts
- code/frontend/src/api/teaching/students.ts
- code/frontend/src/components/contests/ContestOverviewPanel.vue
- code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue
- code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue
- code/frontend/src/features/contest-awd-admin/model/useAwdChallengeLinkOperations.ts
- code/frontend/src/features/contest-awd-admin/model/useAwdContestSnapshotLoader.ts
- code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts
- code/frontend/src/features/contest-detail/model/useContestListPage.ts
- code/frontend/src/features/contest-workbench/model/useContestChallengeOrchestration.ts
- code/frontend/src/features/contest-workbench/model/useContestEditAwdWorkspace.ts
- code/frontend/src/features/platform-contests/ui/ContestOperationsHubWorkspacePanel.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestRulesSection.vue
- code/frontend/src/features/platform-contests/ui/PlatformContestTable.vue
- code/frontend/src/features/scoreboard/model/useScoreboardContestDirectoryPage.ts
- code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts
- code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisDataState.ts
- code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts
- code/frontend/src/features/teacher-workspace/model/useWorkspace.ts
- code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts
- code/frontend/src/views/contests/__tests__/ContestDetail.test.ts
- code/frontend/src/utils/__tests__/skillProfile.test.ts
- code/frontend/src/utils/contest.ts
- code/frontend/src/utils/skillProfile.ts
- code/frontend/src/utils/platformContestAwdChallengeLinks.ts

## After implementation
- `contest` / `skill profile` / `AWD challenge link mapper` 不再伪装成 `utils/*`。
- 这三组能力拥有明确 entity owner。
- 旧 `utils/contest.ts`、`utils/skillProfile.ts`、`utils/platformContestAwdChallengeLinks.ts` 退出代码主路径。
