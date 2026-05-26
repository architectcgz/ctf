import type { ContestChallengeItem, ID } from '@/api/contracts'

export type AWDRuntimeChallenge = ContestChallengeItem & {
  awd_service_id: ID
  awd_challenge_id: ID
}

export function isAwdRuntimeChallenge(
  challenge: ContestChallengeItem
): challenge is AWDRuntimeChallenge {
  return Boolean(challenge.awd_service_id && challenge.awd_challenge_id)
}
