export type StudentDashboardRedirectRouteName = 'TeacherDashboard' | 'PlatformOverview'

export const studentDashboardChallengesRoute = {
  name: 'Challenges',
} as const

export const studentDashboardSkillProfileRoute = {
  name: 'SkillProfile',
} as const

export function studentDashboardRoleRedirectRoute(routeName: StudentDashboardRedirectRouteName) {
  return {
    name: routeName,
  } as const
}

export function studentDashboardCategoryChallengesRoute(category: string) {
  return {
    name: 'Challenges',
    query: {
      category,
    },
  } as const
}

export function studentDashboardDifficultyChallengesRoute(difficulty: string) {
  return {
    name: 'Challenges',
    query: {
      difficulty,
    },
  } as const
}

export function studentDashboardChallengeDetailRoute(challengeId: string) {
  return {
    name: 'ChallengeDetail',
    params: {
      id: challengeId,
    },
  } as const
}
