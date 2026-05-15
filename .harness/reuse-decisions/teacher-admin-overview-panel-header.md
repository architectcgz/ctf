# Reuse Decision

## Change type
- page
- component
- layout

## Existing code searched
- code/frontend/src/style.css
- code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentTimelinePage.vue
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue
- code/frontend/src/components/platform/contest/ContestOrchestrationPage.vue
- code/frontend/src/components/platform/contest/ContestOperationsHubHeroPanel.vue
- code/frontend/src/components/platform/user/UserGovernancePage.vue
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue

## Similar implementations found
- code/frontend/src/style.css
- code/frontend/src/components/dashboard/student/StudentOverviewStyleEditorial.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/components/dashboard/student/StudentCategoryProgressPage.vue
- code/frontend/src/components/dashboard/student/StudentTimelinePage.vue
- code/frontend/src/components/dashboard/student/StudentDifficultyPage.vue

## Decision
extend_existing

## Reason
学生侧已经抽出了 `workspace-panel-header`、`workspace-panel-header__intro`、`workspace-panel-header__actions`、`workspace-panel-header__summary` 和 `workspace-panel-divider` 这一套共享原语。

这次教师/管理员 overview 面板不新建局部头部样式，也不继续占用 `workspace-page-header`。直接复用现有共享结构，把指标卡片并入 `__summary`，需要分隔目录区块时插入 `workspace-panel-divider`，这样可以保持学生、教师、管理员三端的面板节奏一致。

## Files to modify
- code/frontend/src/components/platform/contest/ContestOrchestrationPage.vue
- code/frontend/src/components/platform/contest/ContestOperationsHubHeroPanel.vue
- code/frontend/src/components/platform/user/UserGovernancePage.vue
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue
- code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue
- code/frontend/src/views/platform/__tests__/ContestManage.test.ts
- code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts
- code/frontend/src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts
- code/frontend/src/views/platform/__tests__/UserManage.test.ts
- code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase22.test.ts
- code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts
- code/frontend/src/views/__tests__/journalEyebrowStyles.test.ts
- code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts
- code/frontend/src/views/__tests__/workspaceShellStyles.test.ts
