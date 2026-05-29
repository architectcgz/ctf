export type AwdReviewIndexPageScope = 'teacher' | 'platform'

interface AwdReviewIndexHomeRoute {
  name: 'TeacherDashboard' | 'PlatformOverview'
}

interface AwdReviewDetailRoute {
  name: 'TeacherAWDReviewDetail' | 'PlatformAwdReviewDetail'
  params: {
    contestId: string
  }
}

export function resolveAwdReviewIndexHomeRoute(scope: AwdReviewIndexPageScope): AwdReviewIndexHomeRoute {
  return {
    name: scope === 'teacher' ? 'TeacherDashboard' : 'PlatformOverview',
  }
}

export function buildAwdReviewDetailRoute(
  scope: AwdReviewIndexPageScope,
  contestId: string
): AwdReviewDetailRoute {
  return {
    name: scope === 'teacher' ? 'TeacherAWDReviewDetail' : 'PlatformAwdReviewDetail',
    params: { contestId },
  }
}
