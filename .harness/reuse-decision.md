# Reuse Decision

This file is only for the current task and may be overwritten.
Durable reuse knowledge belongs in `harness/reuse/index.yaml`; append-only summaries belong in `harness/reuse/history.md`.

## Change type
- page
- component
- styling-token

## Existing code searched
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/entities/challenge/ui/ChallengeCategoryPill.vue
- code/frontend/src/entities/challenge/ui/ChallengeDifficultyText.vue
- code/frontend/src/entities/challenge/ui/ChallengeCategoryDifficultyPills.vue
- code/frontend/src/components/teacher/student-management/StudentManagementPage.vue
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/views/instances/InstanceList.vue
- code/frontend/src/components/teacher/TeacherClassReviewPanel.vue
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue
- code/frontend/src/components/platform/writeup/ChallengeWriteupEditorPage.vue
- code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue

## Similar implementations found
- code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue
- code/frontend/src/entities/challenge/ui/ChallengeMetaStrip.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/views/challenges/ChallengeList.vue

## Decision
- refactor_existing

## Reason
- 本次不是新增另一套分类/难度胶囊组件，而是继续扩展 challenge entity 现有 `ChallengeCategoryPill`、`ChallengeDifficultyText`、`ChallengeCategoryDifficultyPills` 和 presentation helper。
- `/academy/students` 的薄弱项列语义上对应题目分类弱项；识别到 `web/pwn/reverse/crypto/misc/forensics` 时应复用题目分类胶囊色，不识别时保留 muted fallback。
- 全站已经有 `--challenge-category-pill-*` 和 `--challenge-difficulty-pill-*`，本次只补齐漏用点，避免页面局部再用 `journal-accent`、`color-success` 或裸文本表达题目分类/难度。
- 回归测试沿用现有页面/抽取测试，补充 source-level 断言锁定共享组件和关键页面的复用关系。

## Files to modify
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/entities/challenge/model/index.ts
- code/frontend/src/entities/challenge/index.ts
- code/frontend/src/entities/challenge/ui/ChallengeCategoryPill.vue
- code/frontend/src/entities/challenge/ui/ChallengeDifficultyText.vue
- code/frontend/src/entities/challenge/ui/ChallengeCategoryDifficultyPills.vue
- code/frontend/src/components/teacher/student-management/StudentManagementPage.vue
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/views/instances/InstanceList.vue
- code/frontend/src/components/teacher/TeacherClassReviewPanel.vue
- code/frontend/src/components/teacher/StudentInsightPanel.vue
- code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue
- code/frontend/src/components/platform/writeup/ChallengeWriteupEditorPage.vue
- code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue
- code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue
- code/frontend/src/components/platform/contest/ContestAwdServiceDirectory.vue
- code/frontend/src/components/teacher/TeacherClassInsightsPanel.vue
- code/frontend/src/components/teacher/TeacherInterventionPanel.vue
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/views/profile/SkillProfile.vue
- code/frontend/src/views/platform/ThemePreview.vue
- code/frontend/src/views/UILab.vue
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts

## After implementation
- Implemented by extending the existing challenge entity presentation helpers and shared pill components.
- Verified with targeted Vitest coverage and `vue-tsc --noEmit`.
