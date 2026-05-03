import { type ComputedRef, type Ref } from 'vue'

import {
  getChallengeDetail,
  getCommunityChallengeSolutions,
  getRecommendedChallengeSolutions,
} from '@/api/challenge'
import type {
  ChallengeDetailData,
  CommunityChallengeSolutionData,
  RecommendedChallengeSolutionData,
} from '@/api/contracts'

interface UseChallengeDetailDataLoaderOptions {
  challengeId: ComputedRef<string>
  challenge: Ref<ChallengeDetailData | null>
  loading: Ref<boolean>
  recommendedSolutions: Ref<RecommendedChallengeSolutionData[]>
  communitySolutions: Ref<CommunityChallengeSolutionData[]>
  onSolutionsLoadFailed: () => void
  onChallengeLoadFailed: () => void
}

export function useChallengeDetailDataLoader(options: UseChallengeDetailDataLoaderOptions) {
  const {
    challengeId,
    challenge,
    loading,
    recommendedSolutions,
    communitySolutions,
    onSolutionsLoadFailed,
    onChallengeLoadFailed,
  } = options

  let latestChallengeRequestId = 0

  function clearSolutions() {
    recommendedSolutions.value = []
    communitySolutions.value = []
  }

  async function loadSolutions(id: string, requestId = latestChallengeRequestId): Promise<void> {
    try {
      const [recommended, communityPage] = await Promise.all([
        getRecommendedChallengeSolutions(id),
        getCommunityChallengeSolutions(id),
      ])
      if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
        return
      }
      recommendedSolutions.value = recommended
      communitySolutions.value = communityPage.list
    } catch {
      if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
        return
      }
      clearSolutions()
      onSolutionsLoadFailed()
    }
  }

  async function loadChallenge(): Promise<void> {
    const id = challengeId.value
    const requestId = ++latestChallengeRequestId
    loading.value = true

    try {
      const detail = await getChallengeDetail(id)
      if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
        return
      }
      challenge.value = detail

      if (detail.is_solved) {
        await loadSolutions(id, requestId)
      } else {
        clearSolutions()
      }
    } catch {
      if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
        return
      }
      onChallengeLoadFailed()
    } finally {
      if (requestId === latestChallengeRequestId && id === challengeId.value) {
        loading.value = false
      }
    }
  }

  return {
    clearSolutions,
    loadSolutions,
    loadChallenge,
  }
}
