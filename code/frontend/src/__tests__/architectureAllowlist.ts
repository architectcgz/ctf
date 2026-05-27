export const viewLineLimit = 500

export const oversizedViewAllowlist = new Set<string>([])

export const componentFeatureImportAllowlist = new Set([
  'components/challenge/ChallengeWorkspaceShell.vue -> @/features/challenge-detail',
  'components/challenge/ChallengeSolutionsPanel.vue -> @/features/challenge-detail',
  'components/challenge/ChallengeSubmissionRecordsPanel.vue -> @/features/challenge-detail',
  'components/contests/ContestAWDWorkspacePanel.vue -> @/features/contest-awd-workspace',
  'components/contests/ContestAnnouncementRealtimeBridge.vue -> @/features/contest-announcements',
  'components/contests/awd/AWDDefenseColumn.vue -> @/features/contest-awd-workspace',
  'components/contests/awd/AWDDefenseOperationsPanel.vue -> @/features/contest-awd-workspace',
  'components/contests/awd/AWDDefenseServiceList.vue -> @/features/contest-awd-workspace',
  'components/layout/AppLayout.vue -> @/features/notifications',
  'components/layout/NotificationDrawer.vue -> @/features/notifications',
  'components/layout/TopNav.vue -> @/features/auth',
  'components/notifications/AdminNotificationPublishDrawer.vue -> @/features/admin-notification-publisher',
  'components/platform/challenge/ChallengeManageDirectoryPanel.vue -> @/features/platform-challenges',
  'components/platform/contest/AWDChallengeConfigPanel.vue -> @/features/awd-inspector',
  'components/platform/contest/AWDOperationsPanel.vue -> @/features/contest-awd-admin',
  'components/platform/contest/AWDRoundInspector.vue -> @/features/awd-inspector',
  'components/platform/contest/AWDTrafficPanel.vue -> @/features/awd-inspector',
  'components/platform/contest/ContestAnnouncementManageDrawer.vue -> @/features/contest-announcements',
  'components/platform/contest/ContestChallengeFilterStrip.vue -> @/features/contest-workbench',
  'components/platform/contest/ContestChallengeOrchestrationPanel.vue -> @/features/contest-workbench',
  'components/platform/contest/ContestEditWorkspacePanel.vue -> @/features/contest-workbench',
  'components/platform/contest/ContestEditWorkspacePanel.vue -> @/features/platform-contests',
  'components/platform/contest/ContestWorkbenchStageTabs.vue -> @/features/contest-workbench',
  'components/platform/contest/ContestWorkbenchSummaryStrip.vue -> @/features/contest-workbench',
  'components/platform/contest/PlatformContestFormDialog.vue -> @/features/platform-contests',
  'components/platform/contest/PlatformContestFormPanel.vue -> @/features/platform-contests',
  'components/platform/contest/awdInspector.types.ts -> @/features/awd-inspector',
  'components/platform/user/PlatformUserFormDialog.vue -> @/features/platform-user-management',
  'components/platform/topology/TopologyCanvasBoard.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyConnectivitySections.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyNetworkSection.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyNodeEditor.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyNodeSection.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyTemplateWorkbench.vue -> @/features/challenge-topology-studio/model',
  'components/scoreboard/ScoreboardRealtimeBridge.vue -> @/features/scoreboard',
  'components/teacher/InterventionPanel.vue -> @/features/teacher-student-analysis',
  'components/teacher/reports/ClassReportExportDialog.vue -> @/features/teacher-class-report-export',
])

export const widgetLegacyComponentImportAllowlist = new Set<string>([])

export const componentNonContractApiAllowlist = new Set([
  'components/teacher/StudentInsightPanel.vue -> @/api/teacher',
  'components/teacher/student-insight/StudentInsightAttackSessionsSection.vue -> @/api/teacher',
])

export const widgetNonContractApiAllowlist = new Set([
  'widgets/teacher-student-review-workspace/StudentReviewWorkspace.vue -> @/api/teacher',
])

export const commonForbiddenImportAllowlist = new Set([
  'components/common/InstancePanel.vue -> @/api/contracts',
  'entities/challenge/model/presentation.ts -> @/api/contracts',
  'entities/challenge/ui/ChallengeCategoryDifficultyPills.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeCategoryPill.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeCategoryText.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeDifficultyText.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeDirectoryRow.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeMetaStrip.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeProfileMetaGrid.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeProfileSummaryStrip.vue -> @/api/contracts',
])

export const legacyComponentPageAllowlist = new Set<string>([])

export const featureRouterImportAllowlist = new Set([
  'features/audit-log/model/useAuditLogPage.ts -> vue-router',
  'features/auth/model/useAuth.ts -> vue-router',
  'features/auth/model/useLoginViewPage.ts -> @/router/guards',
  'features/auth/model/useLoginViewPage.ts -> vue-router',
  'features/challenge-detail/model/useChallengeDetailPage.ts -> vue-router',
  'features/challenge-list/model/useChallengeListPage.ts -> vue-router',
  'features/challenge-package-import/model/useChallengeImportManagePage.ts -> vue-router',
  'features/challenge-package-import/model/useChallengeImportPreviewPage.ts -> vue-router',
  'features/challenge-package-import/model/useChallengePackageFormatPage.ts -> vue-router',
  'features/contest-awd-config/model/useAwdChallengeSelection.ts -> vue-router',
  'features/contest-awd-config/model/useContestAwdConfigPage.ts -> vue-router',
  'features/contest-detail/model/useContestDetailRoutePage.ts -> vue-router',
  'features/contest-detail/model/useContestListPage.ts -> vue-router',
  'features/notifications/model/useNotificationDetailPage.ts -> vue-router',
  'features/notifications/model/useNotificationDrawer.ts -> vue-router',
  'features/notifications/model/useNotificationListPage.ts -> vue-router',
  'features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts -> vue-router',
  'features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts -> vue-router',
  'features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts -> vue-router',
  'features/platform-challenge-detail/model/usePlatformChallengeDetailRoutePage.ts -> vue-router',
  'features/platform-challenges/model/useChallengeManagePage.ts -> vue-router',
  'features/platform-challenges/model/useChallengeManagePresentation.ts -> vue-router',
  'features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts -> vue-router',
  'features/platform-challenges/model/useChallengeWriteupPage.ts -> vue-router',
  'features/platform-challenges/model/useChallengeWriteupViewPage.ts -> vue-router',
  'features/platform-user-management/model/useUserGovernancePanelRoute.ts -> vue-router',
  'features/platform-contests/model/useContestAnnouncementsPage.ts -> vue-router',
  'features/platform-contests/model/useContestManagePage.ts -> vue-router',
  'features/platform-contests/model/useContestEditPage.ts -> vue-router',
  'features/platform-contests/model/useContestOperationsHubPage.ts -> vue-router',
  'features/platform-contests/model/useContestOperationsPage.ts -> vue-router',
  'features/platform-overview/model/useCheatDetectionPage.ts -> vue-router',
  'features/platform-overview/model/usePlatformOverviewPage.ts -> vue-router',
  'features/class-workspace-redirect/model/useClassWorkspaceSection.ts -> vue-router',
  'features/class-students-workspace/model/useClassStudentsPage.ts -> vue-router',
  'features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts -> vue-router',
  'features/student-analysis-workspace/model/useStudentAnalysisPage.ts -> vue-router',
  'features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts -> vue-router',
  'features/student-review-archive-workspace/model/useStudentReviewArchivePage.ts -> vue-router',
  'features/platform-class-management/model/usePlatformClassManagementPage.ts -> vue-router',
  'features/platform-instance-management/model/usePlatformInstanceManagementPage.ts -> vue-router',
  'features/platform-student-management/model/usePlatformStudentManagementPage.ts -> vue-router',
  'features/scoreboard/model/useScoreboardDetailPage.ts -> vue-router',
  'features/scoreboard/model/useScoreboardRoutePage.ts -> vue-router',
  'features/skill-profile/model/useSkillProfilePage.ts -> vue-router',
  'features/student-dashboard/model/useStudentDashboardData.ts -> vue-router',
  'features/student-dashboard/model/useStudentDashboardPage.ts -> vue-router',
  'features/awd-review-workspace/model/useAwdReviewIndex.ts -> vue-router',
  'features/teacher-class-management/model/useClassManagementPage.ts -> vue-router',
  'features/teacher-dashboard/model/useDashboardPage.ts -> vue-router',
  'features/teacher-instances/model/useInstanceManagementPage.ts -> vue-router',
  'features/teacher-student-management/model/useStudentManagementPage.ts -> vue-router',
])

export const utilityBoundaryImportAllowlist = new Set([
  'utils/contest.ts -> @/api/contracts',
  'utils/platformContestAwdChallengeLinks.ts -> @/api/contracts',
  'utils/skillProfile.ts -> @/api/contracts',
])

export const composableMultiBoundaryAllowlist = new Set([
  'composables/useWebSocket.ts -> api+store',
])
