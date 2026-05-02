import { ref, type Ref } from 'vue'

import { submitContestFlag } from '@/api/contest'
import type { ContestChallengeItem, ContestDetailData, SubmitFlagData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseContestFlagSubmissionOptions {
  contest: Ref<ContestDetailData | null>
  challenges: Ref<ContestChallengeItem[]>
  onSelectedChallengeChange?: (challengeId: string | null) => void
}

function buildContestSubmitMessage(result: SubmitFlagData): string {
  if (result.is_correct) {
    return `正确！+${result.points ?? 0} 分`
  }
  return 'Flag 错误，请重试'
}

export function useContestFlagSubmission(options: UseContestFlagSubmissionOptions) {
  const { contest, challenges, onSelectedChallengeChange } = options
  const toast = useToast()

  const selectedChallenge = ref<ContestChallengeItem | null>(null)
  const flagInput = ref('')
  const submitting = ref(false)
  const submitResult = ref<SubmitFlagData | null>(null)

  function clearSubmissionState() {
    flagInput.value = ''
    submitResult.value = null
  }

  function syncSelectedChallengeById(challengeId: string) {
    selectedChallenge.value = challengeId
      ? challenges.value.find((challenge) => challenge.id === challengeId) ?? null
      : null
    clearSubmissionState()
  }

  function selectChallenge(challenge: ContestChallengeItem) {
    selectedChallenge.value = challenge
    clearSubmissionState()
    onSelectedChallengeChange?.(challenge.id)
  }

  async function submitFlagAction() {
    if (submitting.value) {
      return
    }

    const flag = flagInput.value.trim()
    if (!flag) {
      toast.warning('请输入 Flag')
      return
    }
    if (flag.length < 5 || flag.length > 200) {
      toast.warning('Flag 长度应在 5-200 字符之间')
      return
    }
    if (!selectedChallenge.value || !contest.value) {
      return
    }

    submitting.value = true
    submitResult.value = null

    try {
      const result = await submitContestFlag(contest.value.id, selectedChallenge.value.id, flag)
      submitResult.value = {
        ...result,
        message: buildContestSubmitMessage(result),
      }

      if (result.is_correct) {
        const solvedChallengeId = selectedChallenge.value.id
        challenges.value = challenges.value.map((challenge) =>
          challenge.id === solvedChallengeId ? { ...challenge, is_solved: true } : challenge
        )
        selectedChallenge.value = {
          ...selectedChallenge.value,
          is_solved: true,
        }
        flagInput.value = ''
      }
    } catch (error) {
      console.error(error)
      toast.error(error instanceof Error ? error.message : '提交失败，请稍后重试')
    } finally {
      submitting.value = false
    }
  }

  return {
    selectedChallenge,
    flagInput,
    submitting,
    submitResult,
    clearSubmissionState,
    syncSelectedChallengeById,
    selectChallenge,
    submitFlagAction,
  }
}
