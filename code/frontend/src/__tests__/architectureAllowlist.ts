export const viewLineLimit = 500

export const featureRouterImportAllowlist = new Set([
  'features/auth/model/useLoginPage.ts -> @/router/guards',
  'features/auth/model/useLoginPage.ts -> vue-router',
  'features/challenge-detail/model/useChallengeDetailPage.ts -> vue-router',
  'features/contest-awd-config/model/useContestAwdConfigPage.ts -> vue-router',
  'features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts -> vue-router',
])
