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
  solutionsLoading: Ref<boolean>
  recommendedSolutions: Ref<RecommendedChallengeSolutionData[]>
  communitySolutions: Ref<CommunityChallengeSolutionData[]>
  onSolutionsLoadFailed: () => void
  onChallengeLoadFailed: (error: unknown) => void
}

export function useChallengeDetailDataLoader(options: UseChallengeDetailDataLoaderOptions) {
  const {
    challengeId,
    challenge,
    loading,
    solutionsLoading,
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

  async function loadSolutions(id: string, requestId = latestChallengeRequestId): Promise<boolean> {
    solutionsLoading.value = true
    try {
      const [recommended, communityPage] = await Promise.all([
        getRecommendedChallengeSolutions(id),
        getCommunityChallengeSolutions(id),
      ])
      if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
        return false
      }
      recommendedSolutions.value = recommended
      communitySolutions.value = communityPage.list
      return true
    } catch {
      if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
        return false
      }
      clearSolutions()
      onSolutionsLoadFailed()
      return false
    } finally {
      if (requestId === latestChallengeRequestId && id === challengeId.value) {
        solutionsLoading.value = false
      }
    }
  }

  async function loadChallenge(): Promise<boolean> {
    const id = challengeId.value
    const requestId = ++latestChallengeRequestId
    loading.value = true

    try {
      const detail = await getChallengeDetail(id)
      if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
        return false
      }
      challenge.value = detail
      clearSolutions()
      return true
    } catch (error) {
      if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
        return false
      }
      onChallengeLoadFailed(error)
      return false
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
