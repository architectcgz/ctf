# Reuse Decision

## Change type
composition / layout / shared-state / docs / test

## Existing code searched
- `code/frontend/src/composables/useWorkspaceShellNavigation.ts`
- `code/frontend/src/composables/useBackofficeBreadcrumbDetail.ts`
- `code/frontend/src/composables/useToast.ts`
- `code/frontend/src/composables/useDestructiveConfirm.ts`
- `code/frontend/src/shared/ui/layout/**`
- `code/frontend/src/shared/ui/common/AppToast.vue`
- `code/frontend/src/shared/ui/common/AppDestructiveConfirm.vue`
- `code/frontend/src/features/**`
- `code/frontend/src/pages/**`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/06-components.md`

## Similar implementations found
- `shared/model/layout/*` 已经承接 layout 相关 bridge 和 view-state，说明 layout 壳依赖的共享 composition 适合继续收口到 `shared/model/layout/*`
- `shared/lib/navigation/routeTarget.ts` 已作为共享导航薄契约落到 `shared/lib/*`，说明共享运行时 / contract 层不再需要继续挂在 `src/composables`
- `shared/ui/common/AppToast.vue`、`shared/ui/common/AppDestructiveConfirm.vue` 已经分别承接 toast 与 destructive confirm 的可渲染表面，状态 owner 再留在 `src/composables` 会形成“UI 已 shared，state 仍历史目录”的半迁移状态

## Decision
refactor_existing

## Reason
- `src/composables` 目前同时承载共享 UI 状态、layout owner、router transport 和局部业务能力，已经成为历史杂物层
- `useWorkspaceShellNavigation`、`useBackofficeBreadcrumbDetail` 明确服务 `shared/ui/layout/*`，应迁到 `shared/model/layout/*`
- `useToast`、`useDestructiveConfirm` 是共享 UI 原语对应的状态 owner，应迁到 `shared/model/common/*`
- 本轮只收口边界最清楚的第一批 4 个 composables，不把 `routeQueryTransport`、`routeNavigationTransport`、`useRouteQueryTabs` 一起卷入，避免一次触碰过多 route-aware owner

## Files to modify
- `.harness/reuse-decisions/composables-owner-cleanup-batch1.md`
- `docs/plan/impl-plan/2026-05-31-composables-owner-cleanup-batch1-plan.md`
- `docs/architecture/README.md`
- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/api/__tests__/request.test.ts`
- `code/frontend/src/composables/useClipboard.ts`
- `code/frontend/src/composables/useWorkspaceShellNavigation.ts`
- `code/frontend/src/composables/useBackofficeBreadcrumbDetail.ts`
- `code/frontend/src/composables/useToast.ts`
- `code/frontend/src/composables/useDestructiveConfirm.ts`
- `code/frontend/src/composables/__tests__/useWorkspaceShellNavigation.test.ts`
- `code/frontend/src/composables/useDestructiveConfirm.test.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.test.ts`
- `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
- `code/frontend/src/features/admin-notification-publisher/ui/__tests__/AdminNotificationPublishDrawer.test.ts`
- `code/frontend/src/features/auth/model/useAuth.test.ts`
- `code/frontend/src/features/auth/model/useAuth.ts`
- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewExportFlow.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailInteractions.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.test.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeInstance.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeWriteupSubmissionFlow.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyCanvasActions.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyPersistenceActions.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyStructureMutations.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyTemplateApply.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useTopologyTemplateMutations.ts`
- `code/frontend/src/features/challenge-topology-studio/ui/ChallengeTopologyStudioPage.test.ts`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupEditorPage.ts`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupManagePanel.test.ts`
- `code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupPages.test.ts`
- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementManagement.test.ts`
- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementManagement.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdChallengeLinkOperations.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdReadinessDecision.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdRoundOperations.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdServiceOperations.ts`
- `code/frontend/src/features/contest-awd-admin/model/usePlatformContestAwd.test.ts`
- `code/frontend/src/features/contest-awd-config/model/useAwdCheckerSaveFlow.ts`
- `code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.test.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAccessActions.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackSubmission.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceServiceActions.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useContestAWDWorkspace.test.ts`
- `code/frontend/src/features/contest-detail/model/useContestDetailPage.ts`
- `code/frontend/src/features/contest-detail/model/useContestFlagSubmission.ts`
- `code/frontend/src/features/contest-detail/model/useContestTeamActions.ts`
- `code/frontend/src/features/contest-projector/model/useContestProjectorData.test.ts`
- `code/frontend/src/features/contest-projector/model/useContestProjectorData.ts`
- `code/frontend/src/features/contest-projector/model/useContestProjectorPage.ts`
- `code/frontend/src/features/contest-workbench/model/useContestChallengeMutations.ts`
- `code/frontend/src/features/contest-workbench/model/useContestChallengeOrchestration.ts`
- `code/frontend/src/features/contest-workbench/model/useContestEditAwdWorkspace.ts`
- `code/frontend/src/features/contest-workbench/model/useContestExportFlow.ts`
- `code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.test.ts`
- `code/frontend/src/features/image-management/model/useImageManageMutations.ts`
- `code/frontend/src/features/instance-list/model/useInstanceListPage.ts`
- `code/frontend/src/features/instance-list/model/useInstanceOperations.ts`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationDrawer.ts`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationRealtime.test.ts`
- `code/frontend/src/features/notifications/model/useNotificationRealtime.ts`
- `code/frontend/src/features/platform/awd-challenges/model/usePlatformAwdChallenges.test.ts`
- `code/frontend/src/features/platform/awd-challenges/model/usePlatformAwdChallenges.ts`
- `code/frontend/src/features/platform/challenge-detail/model/usePlatformChallengeDetailPage.ts`
- `code/frontend/src/features/platform/challenge-package-import/model/useChallengePackageImport.test.ts`
- `code/frontend/src/features/platform/challenge-package-import/model/useChallengePackageImport.ts`
- `code/frontend/src/features/platform/challenges/model/usePlatformChallenges.test.ts`
- `code/frontend/src/features/platform/challenges/model/usePlatformChallenges.ts`
- `code/frontend/src/features/platform/contests/model/useContestEditPage.ts`
- `code/frontend/src/features/platform/contests/model/useContestOperationsPage.ts`
- `code/frontend/src/features/platform/contests/model/usePlatformContests.ts`
- `code/frontend/src/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUserManagePage.ts`
- `code/frontend/src/features/platform/user-management/model/usePlatformUsers.ts`
- `code/frontend/src/features/profile/model/useSecuritySettingsPage.ts`
- `code/frontend/src/features/scoreboard/model/useContestScoreboardRealtime.test.ts`
- `code/frontend/src/features/scoreboard/model/useContestScoreboardRealtime.ts`
- `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`
- `code/frontend/src/features/scoreboard/model/useScoreboardView.ts`
- `code/frontend/src/features/teacher/instances/model/useInstanceManagementPage.ts`
- `code/frontend/src/features/teacher/instances/model/useInstances.ts`
- `code/frontend/src/features/teaching/class-report-export/model/useClassReportExport.ts`
- `code/frontend/src/features/teaching/student-analysis-review/model/useReviewArchiveExportFlow.ts`
- `code/frontend/src/features/teaching/student-analysis-review/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/teaching/student-review-archive-workspace/model/useStudentReviewArchivePage.ts`
- `code/frontend/src/features/teaching/student-review-archive/model/useStudentReviewArchive.ts`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/pages/notifications/__tests__/NotificationDetail.test.ts`
- `code/frontend/src/pages/notifications/__tests__/NotificationList.test.ts`
- `code/frontend/src/pages/platform/__tests__/ImageManage.test.ts`
- `code/frontend/src/pages/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/pages/platform/__tests__/UserManage.test.ts`
- `code/frontend/src/pages/platform/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAnnouncements.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAwdConfig.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestEdit.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestProjector.test.ts`
- `code/frontend/src/pages/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/router/__tests__/guards.test.ts`
- `code/frontend/src/router/guards.ts`
- `code/frontend/src/shared/model/common/useDestructiveConfirm.test.ts`
- `code/frontend/src/shared/model/common/useDestructiveConfirm.ts`
- `code/frontend/src/shared/model/common/useToast.ts`
- `code/frontend/src/shared/model/layout/__tests__/useWorkspaceShellNavigation.test.ts`
- `code/frontend/src/shared/model/layout/index.ts`
- `code/frontend/src/shared/model/layout/sidebar/useSidebarNavigationViewState.ts`
- `code/frontend/src/shared/model/layout/topnav/useTopNavViewState.ts`
- `code/frontend/src/shared/model/layout/useBackofficeBreadcrumbDetail.ts`
- `code/frontend/src/shared/model/layout/useWorkspaceShellNavigation.ts`
- `code/frontend/src/shared/ui/common/AppDestructiveConfirm.vue`
- `code/frontend/src/shared/ui/common/AppToast.vue`
- `code/frontend/src/shared/ui/common/__tests__/AppToast.test.ts`
- `code/frontend/src/shared/ui/layout/__tests__/NotificationDrawer.test.ts`
- `code/frontend/src/shared/ui/layout/__tests__/TopNav.test.ts`
- `code/frontend/src/shared/ui/layout/topnav/TopNavBreadcrumbs.vue`

## After implementation
- layout 共享壳依赖的 composition owner 将进入 `shared/model/layout/*`
- toast / destructive confirm 的共享状态 owner 将进入 `shared/model/common/*`
- `src/composables` 将减少一批已经有清晰 shared owner 的历史文件
