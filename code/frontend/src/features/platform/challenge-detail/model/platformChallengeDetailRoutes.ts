export const platformChallengeListRoute = {
  name: 'ChallengeManage',
} as const

export function platformChallengeTopologyStudioRoute(challengeId: string) {
  return {
    name: 'PlatformChallengeTopologyStudio',
    params: {
      id: challengeId,
    },
  } as const
}

export function platformChallengeWriteupRoute(challengeId: string, mode: 'view' | 'edit') {
  return {
    name: mode === 'view' ? 'PlatformChallengeWriteupView' : 'PlatformChallengeWriteup',
    params: {
      id: challengeId,
    },
  } as const
}
