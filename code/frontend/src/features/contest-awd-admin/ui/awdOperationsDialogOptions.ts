import type { AdminContestChallengeViewData } from '@/api/contracts'

export function sortAwdChallengeLinks(
  challengeLinks: AdminContestChallengeViewData[]
): AdminContestChallengeViewData[] {
  return [...challengeLinks].sort(
    (a, b) => a.order - b.order || Number(a.challenge_id) - Number(b.challenge_id)
  )
}

export function formatAwdChallengeLabel(challenge: AdminContestChallengeViewData): string {
  const prefix = challenge.title?.trim()
    ? challenge.title.trim()
    : `Challenge #${challenge.challenge_id}`
  return `${prefix} · ${challenge.is_visible ? '可见' : '隐藏'}`
}

export function resolveAwdServiceId(
  challengeLinks: AdminContestChallengeViewData[],
  challengeId: string
): number | null {
  const challenge = challengeLinks.find((item) => item.challenge_id === challengeId)
  if (!challenge?.awd_service_id) {
    return null
  }
  return Number(challenge.awd_service_id)
}
