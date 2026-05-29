export const viewLineLimit = 500

export const featureRouterImportAllowlist = new Set([
  'features/auth/model/useLoginPage.ts -> @/router/guards',
  'features/auth/model/useLoginPage.ts -> vue-router',
  'features/challenge-detail/model/useChallengeDetailPage.ts -> vue-router',
  'features/contest-awd-config/model/useContestAwdConfigPage.ts -> vue-router',
  'features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts -> vue-router',
  'features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts -> vue-router',
  'features/platform-challenges/model/useChallengeManagePage.ts -> vue-router',
  'features/platform-challenges/model/usePlatformChallengeRoutePage.ts -> vue-router',
  'features/platform-user-management/model/usePlatformUserManagePage.ts -> vue-router',
  'features/class-students-workspace/model/useClassStudentsPage.ts -> vue-router',
  'features/student-analysis-workspace/model/useStudentAnalysisPage.ts -> vue-router',
])
