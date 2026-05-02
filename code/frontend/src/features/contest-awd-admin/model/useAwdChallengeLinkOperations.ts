import { ref, type Ref } from 'vue'

import {
  createContestAWDService,
  listContestAWDServices,
  updateContestAWDService,
} from '@/api/admin/contests'
import type {
  AdminContestChallengeData,
  ContestDetailData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import { mapPlatformContestAwdServicesToChallengeLinks } from '@/utils/platformContestAwdChallengeLinks'

interface UseAwdChallengeLinkOperationsOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
  onAfterMutate: () => Promise<void>
}

export function useAwdChallengeLinkOperations(options: UseAwdChallengeLinkOperationsOptions) {
  const { selectedContest, onAfterMutate } = options
  const toast = useToast()

  const challengeLinks = ref<AdminContestChallengeData[]>([])
  const savingChallengeConfig = ref(false)

  async function refreshChallengeLinks() {
    if (!selectedContest.value) {
      challengeLinks.value = []
      return
    }
    const nextServices = await listContestAWDServices(selectedContest.value.id)
    challengeLinks.value = mapPlatformContestAwdServicesToChallengeLinks(nextServices)
  }

  async function createChallengeLink(payload: {
    challenge_id: number
    awd_challenge_id?: number
    points: number
    order?: number
    is_visible?: boolean
    awd_checker_type?: AdminContestChallengeData['awd_checker_type']
    awd_checker_config?: Record<string, unknown>
    awd_sla_score?: number
    awd_defense_score?: number
    awd_checker_preview_token?: string
  }) {
    if (!selectedContest.value) {
      return
    }
    if (!payload.awd_challenge_id) {
      toast.error('请选择 AWD 题目')
      return
    }

    savingChallengeConfig.value = true
    try {
      await createContestAWDService(selectedContest.value.id, {
        awd_challenge_id: payload.awd_challenge_id,
        points: payload.points,
        order: payload.order,
        is_visible: payload.is_visible,
        checker_type: payload.awd_checker_type,
        checker_config: payload.awd_checker_config,
        awd_sla_score: payload.awd_sla_score,
        awd_defense_score: payload.awd_defense_score,
        awd_checker_preview_token: payload.awd_checker_preview_token,
      })
      toast.success('赛事题目已关联')
      await refreshChallengeLinks()
      await onAfterMutate()
    } finally {
      savingChallengeConfig.value = false
    }
  }

  async function updateChallengeLink(
    challengeId: string,
    payload: {
      awd_challenge_id?: number
      points?: number
      order?: number
      is_visible?: boolean
      awd_checker_type?: AdminContestChallengeData['awd_checker_type']
      awd_checker_config?: Record<string, unknown>
      awd_sla_score?: number
      awd_defense_score?: number
      awd_checker_preview_token?: string
    }
  ) {
    if (!selectedContest.value) {
      return
    }

    savingChallengeConfig.value = true
    try {
      const currentChallenge = challengeLinks.value.find((item) => item.challenge_id === challengeId)
      const currentAWDChallengeID = Number(currentChallenge?.awd_challenge_id || 0) || undefined
      const awdChallengeID = payload.awd_challenge_id ?? currentAWDChallengeID
      const points = payload.points ?? currentChallenge?.points
      const order = payload.order ?? currentChallenge?.order
      const isVisible = payload.is_visible ?? currentChallenge?.is_visible

      if (awdChallengeID && points !== undefined) {
        const nextPayload = {
          awd_challenge_id: awdChallengeID,
          points,
          order,
          is_visible: isVisible,
          checker_type: payload.awd_checker_type,
          checker_config: payload.awd_checker_config,
          awd_sla_score: payload.awd_sla_score,
          awd_defense_score: payload.awd_defense_score,
          awd_checker_preview_token: payload.awd_checker_preview_token,
        }
        if (currentChallenge?.awd_service_id) {
          await updateContestAWDService(
            selectedContest.value.id,
            currentChallenge.awd_service_id,
            nextPayload
          )
        } else {
          await createContestAWDService(selectedContest.value.id, nextPayload)
        }
      }
      toast.success('题目配置已更新')
      await refreshChallengeLinks()
      await onAfterMutate()
    } finally {
      savingChallengeConfig.value = false
    }
  }

  return {
    challengeLinks,
    savingChallengeConfig,
    refreshChallengeLinks,
    createChallengeLink,
    updateChallengeLink,
  }
}
