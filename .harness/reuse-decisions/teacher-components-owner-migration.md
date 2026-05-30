# Reuse Decision

## Change type
component / composition / page

## Existing code searched
- `code/frontend/src/components/teacher/*`
- `code/frontend/src/features/teacher-dashboard/*`
- `code/frontend/src/features/teacher-instances/*`
- `code/frontend/src/features/teacher-class-management/*`
- `code/frontend/src/features/class-students-workspace/*`
- `code/frontend/src/features/student-analysis-workspace/*`
- `code/frontend/src/features/teacher-class-report-export/*`
- `code/frontend/src/pages/teacher/__tests__/*`
- `code/frontend/src/pages/__tests__/*`

## Similar implementations found
- `code/frontend/src/features/platform/overview/ui/*`
- `code/frontend/src/features/platform/class-management/ui/*`
- `code/frontend/src/features/platform/student-management/ui/*`
- `code/frontend/src/features/platform/instance-management/ui/*`

## Decision
refactor_existing

## Reason
teacher 侧旧 `components/teacher/*` 目录里的 dashboard、instance-management、class-management、student-insight 组件已经只服务明确 feature owner，但这些 owner 还分成两类：

- 教师独占能力：应收进 `features/teacher/**`
- teacher / platform 共用能力：应收进中性 `features/teaching/**`

因此本轮不新建 bridge，也不保留中间壳，直接把这些 UI 迁到最终 owner 路径，并同步更新 feature/page/test 引用。

`awd-review` 与 `review-archive` 仍由 widget workspace 共同消费，本轮不一并混入，避免把 feature owner 迁移和 workspace owner 迁移耦在一个提交里。

## Files to modify
- `.harness/reuse-decisions/teacher-components-owner-migration.md`
- `docs/plan/impl-plan/2026-05-30-teacher-components-owner-migration-plan.md`
- `code/frontend/src/features/teacher/class-management/index.ts`
- `code/frontend/src/features/teacher/class-management/model/index.ts`
- `code/frontend/src/features/teacher/class-management/model/teacherClassManagementRoutes.ts`
- `code/frontend/src/features/teacher/class-management/model/useClassManagementPage.ts`
- `code/frontend/src/features/teacher/class-management/ui/ClassManagementPage.vue`
- `code/frontend/src/features/teacher/class-management/ui/TeacherClassManagementHeaderActions.vue`
- `code/frontend/src/features/teacher/class-management/ui/TeacherClassManagementRowLink.vue`
- `code/frontend/src/features/teacher/class-management/ui/index.ts`
- `code/frontend/src/features/teacher/dashboard/index.ts`
- `code/frontend/src/features/teacher/dashboard/model/index.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardInsightBuilders.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardOverviewBuilders.ts`
- `code/frontend/src/features/teacher/dashboard/model/teacherDashboardRoutes.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardMetrics.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardMetricsBoundary.test.ts`
- `code/frontend/src/features/teacher/dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardInterventionPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPortraitPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardReviewPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardStudentInsightPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardTrendPanel.vue`
- `code/frontend/src/features/teacher/dashboard/ui/index.ts`
- `code/frontend/src/features/teacher/instances/index.ts`
- `code/frontend/src/features/teacher/instances/model/index.ts`
- `code/frontend/src/features/teacher/instances/model/teacherInstanceManagementRoutes.ts`
- `code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceManagementPage.vue`
- `code/frontend/src/features/teacher/instances/ui/index.ts`
- `code/frontend/src/features/teacher/student-management/index.ts`
- `code/frontend/src/features/teacher/student-management/model/index.ts`
- `code/frontend/src/features/teacher/student-management/model/teacherStudentManagementRoutes.ts`
- `code/frontend/src/features/teacher/student-management/model/useStudentManagementPage.ts`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `code/frontend/src/features/teacher/student-management/ui/index.ts`
- `code/frontend/src/features/teaching/class-insight-window/index.ts`
- `code/frontend/src/features/teaching/class-insight-window/model/index.ts`
- `code/frontend/src/features/teaching/class-insight-window/model/window.ts`
- `code/frontend/src/features/teaching/class-report-export/index.ts`
- `code/frontend/src/features/teaching/class-report-export/model/index.ts`
- `code/frontend/src/features/teaching/class-report-export/model/useClassReportExport.ts`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportContextSection.vue`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportDialog.vue`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportTaskRail.vue`
- `code/frontend/src/features/teaching/class-report-export/ui/classReportExportDialog.css`
- `code/frontend/src/features/teaching/class-report-export/ui/index.ts`
- `code/frontend/src/features/teaching/class-students-workspace/index.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/classStudentsRoutes.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/index.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassStudentsPage.ts`
- `code/frontend/src/features/teaching/class-students-workspace/model/useClassWorkspaceSection.ts`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassInsightsPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassReviewPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsInsightWindowPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassTrendPanel.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/model/index.ts`
- `code/frontend/src/features/teaching/student-analysis-review/model/useInterventionRecommendations.ts`
- `code/frontend/src/features/teaching/student-analysis-review/model/useReviewArchiveExportFlow.ts`
- `code/frontend/src/features/teaching/student-analysis-review/model/useReviewWorkspace.ts`
- `code/frontend/src/features/teaching/student-analysis-review/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/features/teaching/student-analysis-review/ui/InterventionPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-review/ui/index.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/index.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/index.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/studentAnalysisRoutes.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisDataState.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightManualReviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentInsightWriteupsSection.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentReviewWorkspace.test.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentReviewWorkspace.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/index.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/studentInsightShared.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/studentReviewWorkspacePresentation.test.ts`
- `code/frontend/src/features/teaching/student-review-archive/index.ts`
- `code/frontend/src/features/teaching/student-review-archive/model/index.ts`
- `code/frontend/src/features/teaching/student-review-archive/model/presentation.test.ts`
- `code/frontend/src/features/teaching/student-review-archive/model/presentation.ts`
- `code/frontend/src/features/teaching/student-review-archive/model/useStudentReviewArchive.ts`
- `code/frontend/src/features/teaching/student-review-archive-workspace/index.ts`
- `code/frontend/src/features/teaching/student-review-archive-workspace/model/index.ts`
- `code/frontend/src/features/teaching/student-review-archive-workspace/model/studentReviewArchiveRoutes.ts`
- `code/frontend/src/features/teaching/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/pages/teacher/ClassManagementRoutePage.vue`
- `code/frontend/src/pages/teacher/InstanceManagementRoutePage.vue`
- `code/frontend/src/pages/teacher/StudentManagementRoutePage.vue`
- `code/frontend/src/pages/teacher/TeacherClassStudentsRoutePage.vue`
- `code/frontend/src/pages/teacher/TeacherDashboardRoutePage.vue`
- `code/frontend/src/pages/teacher/TeacherStudentAnalysisRoutePage.vue`
- `code/frontend/src/pages/platform/PlatformClassStudentsRoutePage.vue`
- `code/frontend/src/pages/platform/PlatformStudentAnalysisRoutePage.vue`
- `code/frontend/src/pages/review-archive/StudentReviewArchiveRoutePage.vue`
- `code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassWorkspaceSection.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherInterventionPanelLayout.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherSharedDirectoryStyles.test.ts`
- `code/frontend/src/pages/teacher/__tests__/teacherSurface.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/pages/review-archive/__tests__/PlatformStudentReviewArchive.test.ts`
- `code/frontend/src/pages/review-archive/__tests__/TeacherStudentReviewArchive.test.ts`
- `code/frontend/src/pages/__tests__/journalEyebrowStyles.test.ts`
- `code/frontend/src/pages/__tests__/pageTabsStyles.test.ts`
- `code/frontend/src/pages/__tests__/sharedPaginationControls.test.ts`
- `code/frontend/src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/pages/__tests__/workspaceShellStyles.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/__tests__/frontendArchitecturePolicy.ts`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/components/teacher/dashboard/*`
- `code/frontend/src/components/teacher/instance-management/*`
- `code/frontend/src/components/teacher/class-management/*`
- `code/frontend/src/components/teacher/student-insight/*`
- `code/frontend/src/components/teacher/ClassInsightsPanel.vue`
- `code/frontend/src/components/teacher/ClassReviewPanel.vue`
- `code/frontend/src/components/teacher/ClassTrendPanel.vue`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`

## After implementation
- 如果 `teacher` 组件迁移模式后续还会重复用于 `components/contests/*`、`components/platform/challenge/*`，再把稳定模式补进 `.harness/reuse-index/`。
