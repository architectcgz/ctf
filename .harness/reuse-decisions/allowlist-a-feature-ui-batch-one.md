# Reuse Decision

## Change type
frontend feature-ui migration / architecture allowlist reduction

## Existing code searched
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components/scoreboard/ScoreboardRealtimeBridge.vue`
- `code/frontend/src/components/notifications/AdminNotificationPublishDrawer.vue`
- `code/frontend/src/components/platform/user/PlatformUserFormDialog.vue`
- `code/frontend/src/components/platform/challenge/ChallengeManageDirectoryPanel.vue`
- `code/frontend/src/components/teacher/InterventionPanel.vue`
- `code/frontend/src/components/teacher/reports/ClassReportExportDialog.vue`
- `code/frontend/src/features/scoreboard/**`
- `code/frontend/src/features/admin-notification-publisher/**`
- `code/frontend/src/features/platform-user-management/**`
- `code/frontend/src/features/platform-challenges/**`
- `code/frontend/src/features/teacher-student-analysis/**`
- `code/frontend/src/features/teacher-class-report-export/**`
- `code/frontend/src/views/challenges/**`
- `code/frontend/src/views/contests/**`
- `code/frontend/src/views/notifications/**`
- `code/frontend/src/views/platform/**`
- `code/frontend/src/views/teacher/**`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`

## Similar implementations found
- `UserGovernancePage.vue`、`PlatformOverviewPage.vue`、`TeacherDashboardPage.vue`、`ClassStudentsPage.vue`、`StudentAnalysisPage.vue` 已按“feature public API 组合 page model 与 page shell”的方式迁出 legacy component page。
- `ClassReportExportDialog` 已有 `components/reports` 中立入口，说明这类共享对话框已经具备 public owner 收口前提。
- 当前 `componentFeatureImportAllowlist` 剩余 22 条里，这 6 条都直接消费单一 feature model / workflow，没有 topology editor 或 layout infra 那样的多边界混合问题，适合做第一批低风险收口。

## Decision
refactor_existing

## Reason
- 这批文件当前主要问题不是行为错误，而是 UI 文件仍停留在 `components/*`，却直接依赖单一 feature owner，导致 `componentFeatureImportAllowlist` 持续承载历史例外。
- 最小正确改动是把这 6 条迁入各自 feature 的 `ui` 目录，并通过 feature public API 暴露给现有 view / page 消费面。
- 本轮不处理 topology / layout / contest-awd workspace 那些需要先重定边界的条目，避免把简单迁移切片做成设计重构。

## Files to modify
- `.harness/reuse-decisions/allowlist-a-feature-ui-batch-one.md`
- `docs/plan/impl-plan/2026-05-28-allowlist-a-feature-ui-batch-one-plan.md`
- `docs/reviews/frontend/2026-05-28-allowlist-a-feature-ui-batch-one-review.md`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components/scoreboard/ScoreboardRealtimeBridge.vue`
- `code/frontend/src/components/notifications/AdminNotificationPublishDrawer.vue`
- `code/frontend/src/components/platform/user/PlatformUserFormDialog.vue`
- `code/frontend/src/components/platform/challenge/ChallengeManageDirectoryPanel.vue`
- `code/frontend/src/components/teacher/InterventionPanel.vue`
- `code/frontend/src/components/teacher/reports/ClassReportExportDialog.vue`
- `code/frontend/src/components/reports/index.ts`
- `code/frontend/src/components/teacher/reports/index.ts`
- `code/frontend/src/features/scoreboard/index.ts`
- `code/frontend/src/features/admin-notification-publisher/index.ts`
- `code/frontend/src/features/platform-user-management/index.ts`
- `code/frontend/src/features/platform-user-management/ui/index.ts`
- `code/frontend/src/features/platform-challenges/index.ts`
- `code/frontend/src/features/teacher-student-analysis/index.ts`
- `code/frontend/src/features/teacher-class-report-export/index.ts`
- `code/frontend/src/features/scoreboard/ui/index.ts`
- `code/frontend/src/features/scoreboard/ui/ScoreboardRealtimeBridge.vue`
- `code/frontend/src/features/admin-notification-publisher/ui/index.ts`
- `code/frontend/src/features/admin-notification-publisher/ui/AdminNotificationPublishDrawer.vue`
- `code/frontend/src/features/platform-user-management/ui/PlatformUserFormDialog.vue`
- `code/frontend/src/features/platform-challenges/ui/index.ts`
- `code/frontend/src/features/platform-challenges/ui/ChallengeManageDirectoryPanel.vue`
- `code/frontend/src/features/teacher-student-analysis/ui/index.ts`
- `code/frontend/src/features/teacher-student-analysis/ui/InterventionPanel.vue`
- `code/frontend/src/features/teacher-class-report-export/ui/index.ts`
- `code/frontend/src/features/teacher-class-report-export/ui/ClassReportExportDialog.vue`
- `code/frontend/src/views/challenges/**`
- `code/frontend/src/views/contests/**`
- `code/frontend/src/views/notifications/**`
- `code/frontend/src/views/platform/**`
- `code/frontend/src/views/teacher/**`
- `code/frontend/src/features/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teacher-dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/views/__tests__/**`
- `code/frontend/src/components/notifications/__tests__/AdminNotificationPublishDrawer.test.ts`
- `code/frontend/src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Additional touched files during implementation
- `code/frontend/src/components.d.ts`
- `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/notifications/__tests__/AdminNotificationPublishDrawer.test.ts`
- `code/frontend/src/components/teacher/ClassInsightsPanel.vue`
- `code/frontend/src/components/teacher/ClassReviewPanel.vue`
- `code/frontend/src/components/teacher/ClassTrendPanel.vue`
- `code/frontend/src/components/teacher/teacher-panel-shell.css`
- `code/frontend/src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- `code/frontend/src/features/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/contest-awd-workspace/index.ts`
- `code/frontend/src/features/contest-awd-workspace/ui/index.ts`
- `code/frontend/src/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/assets/styles/teacher-panel-shell.css`
- `code/frontend/src/views/notifications/NotificationList.vue`
- `code/frontend/src/views/platform/ChallengeManage.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/platform/UserManage.vue`
- `code/frontend/src/views/scoreboard/ScoreboardDetail.vue`
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `code/frontend/src/views/contests/__tests__/contestStudentActionPrimitives.test.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/challengeManageDirectoryExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `code/frontend/src/views/teacher/ClassManagement.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
- `code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherInterventionPanelLayout.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherPanelShellAdoption.test.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## After implementation
- 这 6 条不再继续占用 `componentFeatureImportAllowlist`；同时因 `ScoreboardRealtimeBridge` 的 legacy component consumer 需要同步收口，`ContestAWDWorkspacePanel.vue` 也一并迁入 `features/contest-awd-workspace/ui`。
- 对应 view / feature 消费面改为从各自 feature public API 读取 UI owner。
- `teacher-panel-shell.css` 已从 `components/teacher` 收口到中立共享位置 `assets/styles`，不再让 feature UI 反向引用 legacy component 目录样式。
- `componentFeatureImportAllowlist` 已从 22 条降到 15 条，剩余重点进一步集中到 challenge-detail、contest-awd-workspace 余项、topology、layout 几组。
