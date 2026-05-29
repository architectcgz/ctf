export const challengeListDashboardRoute = {
  name: 'Dashboard',
} as const

export const challengeListSkillProfileRoute = {
  name: 'SkillProfile',
} as const

export function challengeListDetailRoute(challengeId: string) {
  return {
    name: 'ChallengeDetail',
    params: {
      id: challengeId,
    },
  } as const
}
