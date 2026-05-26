import { computed, toValue, type MaybeRefOrGetter } from 'vue'

import type {
  AWDAttackLogData,
  ContestAWDWorkspaceEventDirection,
  ContestChallengeItem,
} from '@/api/contracts'
import { isAwdRuntimeChallenge, type AWDRuntimeChallenge } from './awdChallengeIdentity'

interface UseAwdWorkspacePresentationOptions {
  challenges: MaybeRefOrGetter<ContestChallengeItem[]>
}

export function useAwdWorkspacePresentation(options: UseAwdWorkspacePresentationOptions) {
  const challengeByChallengeId = computed(() => {
    const map = new Map<string, AWDRuntimeChallenge>()
    for (const item of toValue(options.challenges)) {
      if (isAwdRuntimeChallenge(item)) {
        map.set(item.awd_challenge_id, item)
      }
    }
    return map
  })

  const challengeByServiceId = computed(() => {
    const map = new Map<string, AWDRuntimeChallenge>()
    for (const item of toValue(options.challenges)) {
      if (isAwdRuntimeChallenge(item)) {
        map.set(item.awd_service_id, item)
      }
    }
    return map
  })

  function getChallengeTitleForEvent(event: {
    service_id?: string
    awd_challenge_id: string
  }): string {
    if (event.service_id) {
      const matchedByService = challengeByServiceId.value.get(event.service_id)
      if (matchedByService) return matchedByService.title
    }
    return challengeByChallengeId.value.get(event.awd_challenge_id)?.title || event.awd_challenge_id
  }

  function formatAttackResultToast(result: Pick<
    AWDAttackLogData,
    'service_id' | 'awd_challenge_id' | 'is_success' | 'score_gained'
  >): string {
    const challengeTitle = getChallengeTitleForEvent(result)
    if (result.is_success) return `${challengeTitle}: 攻击成功，+${result.score_gained} 分`
    return `${challengeTitle}: 未获取到有效 Flag。`
  }

  function eventDirectionLabel(direction: ContestAWDWorkspaceEventDirection): string {
    return direction === 'attack_out' ? '对外攻击' : '受到攻击'
  }

  function eventResultLabel(success: boolean): string {
    return success ? '成功' : '失败'
  }

  function formatServiceRef(serviceId?: string): string {
    return `服务 #${serviceId || '--'}`
  }

  return {
    getChallengeTitleForEvent,
    formatAttackResultToast,
    eventDirectionLabel,
    eventResultLabel,
    formatServiceRef,
  }
}
