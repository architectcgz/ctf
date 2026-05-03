import { ref, toValue, type MaybeRefOrGetter } from 'vue'

import { submitContestAWDAttack } from '@/api/contest'
import type { AWDAttackLogData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseAwdWorkspaceAttackSubmissionOptions {
  contestId: MaybeRefOrGetter<string>
  refreshAll: () => Promise<void>
  formatAttackResultToast?: (result: AWDAttackLogData) => string
}

export function useAwdWorkspaceAttackSubmission(options: UseAwdWorkspaceAttackSubmissionOptions) {
  const { contestId, refreshAll, formatAttackResultToast } = options
  const toast = useToast()

  const submitResult = ref<AWDAttackLogData | null>(null)
  const submittingKey = ref('')

  async function submitAttack(
    serviceId: string,
    victimTeamId: number,
    flag: string
  ): Promise<AWDAttackLogData | null> {
    const resolvedContestId = toValue(contestId)
    const normalizedFlag = flag.trim()
    if (!resolvedContestId || !victimTeamId || !normalizedFlag) {
      return null
    }
    if (submittingKey.value) {
      return null
    }

    submittingKey.value = `${serviceId}:${victimTeamId}`
    submitResult.value = null

    try {
      const result = await submitContestAWDAttack(resolvedContestId, serviceId, {
        victim_team_id: victimTeamId,
        flag: normalizedFlag,
      })
      submitResult.value = result
      await refreshAll()
      const formattedMessage = formatAttackResultToast?.(result)
      toast.success(
        formattedMessage ||
          (result.is_success ? `攻击成功，+${result.score_gained} 分` : '攻击未命中有效 flag')
      )
      return result
    } catch (err) {
      console.error(err)
      toast.error(err instanceof Error ? err.message : '提交 stolen flag 失败')
      return null
    } finally {
      submittingKey.value = ''
    }
  }

  return {
    submitAttack,
    submitResult,
    submittingKey,
  }
}
