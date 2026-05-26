# Reuse Decision

## Change type
page / component / layout / table / modal / composition

## Existing code searched
- `code/frontend/src/components/platform/user/UserGovernancePage.vue`
- `code/frontend/src/views/platform/UserManage.vue`
- `code/frontend/src/features/platform-user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue`
- `code/frontend/src/views/teacher/TeacherDashboard.vue`
- `code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/components/common/WorkspaceDirectoryToolbar.vue`
- `code/frontend/src/components/common/WorkspaceDataTable.vue`
- `code/frontend/src/components/common/modal-templates/AdminSurfaceModal.vue`
- `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`

## Similar implementations found
- `code/frontend/src/views/teacher/TeacherDashboard.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/components/teacher/ClassTrendPanel.vue`
- `code/frontend/src/components/teacher/ClassReviewPanel.vue`
- `code/frontend/src/components/teacher/ClassInsightsPanel.vue`
- `code/frontend/src/components/teacher/InterventionPanel.vue`
- `code/frontend/src/components/common/WorkspaceDirectoryToolbar.vue`
- `code/frontend/src/components/common/WorkspaceDataTable.vue`
- `code/frontend/src/components/common/modal-templates/AdminSurfaceModal.vue`

## Decision
refactor_existing

## Reason
- 这次不是新增新的 workspace 页面能力，而是收口三个已经过宽的 page 组件。route view、feature composable、共享表格 / toolbar / modal 原语都已存在，最小正确方案是沿着现有 owner 继续拆稳定展示块，而不是新造 page framework。
- `TeacherDashboard.vue`、`TeacherClassStudents.vue`、`PlatformClassStudents.vue` 已经证明 route/page owner 应留在 view + feature composable，子组件只吃 props / emits；本轮应延续这个边界。
- `UserGovernancePage.vue`、`TeacherDashboardPage.vue`、`ClassStudentsPage.vue` 里已有多段可以独立命名的稳定展示区，但这些区目前被揉在一个大文件里。直接重构现有 page 并抽子组件，改动最小，也最符合仓库最近几轮 student-analysis / workspace page owner 收口的方向。

## Files to modify
- `docs/plan/impl-plan/2026-05-26-teacher-platform-workspace-page-decomposition-implementation-plan.md`
- `.harness/reuse-decisions/teacher-platform-workspace-page-decomposition.md`
- `code/frontend/src/components/platform/user/UserGovernancePage.vue`
- `code/frontend/src/components/platform/user/UserGovernanceOverviewPanel.vue`
- `code/frontend/src/components/platform/user/UserGovernanceDetailModal.vue`
- `code/frontend/src/components/platform/user/UserGovernanceImportPanel.vue`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardPortraitPanel.vue`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardStudentInsightPanel.vue`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardTrendPanel.vue`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardReviewPanel.vue`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardInterventionPanel.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsOverviewPanel.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsInsightWindowPanel.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsDirectoryPanel.vue`
- `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/teacher/__tests__/classStudentsPanelExtraction.test.ts`
- `code/frontend/src/views/teacher/__tests__/classManagementTabsAdoption.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherSurface.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
- `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
- `code/frontend/src/views/__tests__/journalNoteStyles.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `code/frontend/src/views/platform/__tests__/platformRootShellCleanup.test.ts`
- `code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts`
- `docs/reviews/frontend/2026-05-26-teacher-platform-workspace-page-decomposition-review.md`

## After implementation
- 如果这些页面拆分后形成稳定命名模式，可以再看是否需要把 “workspace 大 page 子组件切分规则” 补进本地 `.harness/reuse-index/`，让后续同类页面直接复用。
