export const skillProfileChallengesRoute = {
  name: 'Challenges',
} as const

export function skillProfileChallengeDetailRoute(challengeId: string) {
  return {
    name: 'ChallengeDetail',
    params: {
      id: challengeId,
    },
  } as const
}
