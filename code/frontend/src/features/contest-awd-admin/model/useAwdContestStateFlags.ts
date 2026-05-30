import { computed, type Ref } from 'vue'

import type { AWDRoundData, ContestDetailData } from '@/api/contracts'

interface UseAwdContestStateFlagsOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
  selectedRoundId: Readonly<Ref<string | null>>
  selectedRound: Readonly<Ref<AWDRoundData | null>>
}

export function isAwdRuntimeStageStatus(status?: ContestDetailData['status']): boolean {
  return status === 'running' || status === 'frozen' || status === 'ended'
}

export function useAwdContestStateFlags(options: UseAwdContestStateFlagsOptions) {
  const { selectedContest, selectedRoundId, selectedRound } = options

  const hasSelectedContest = computed(
    () => Boolean(selectedContest.value) && selectedContest.value?.mode === 'awd'
  )
  const runtimeStageReady = computed(() => isAwdRuntimeStageStatus(selectedContest.value?.status))
  const canOperateSelectedRound = computed(
    () => hasSelectedContest.value && Boolean(selectedRoundId.value)
  )
  const shouldUseCurrentRoundCheck = computed(
    () => selectedRound.value?.status === 'running' || !selectedRoundId.value
  )
  const shouldAutoRefresh = computed(() => {
    if (!selectedContest.value || selectedContest.value.mode !== 'awd') {
      return false
    }
    if (selectedContest.value.status !== 'running' && selectedContest.value.status !== 'frozen') {
      return false
    }
    return selectedRound.value?.status === 'running'
  })

  return {
    hasSelectedContest,
    runtimeStageReady,
    canOperateSelectedRound,
    shouldUseCurrentRoundCheck,
    shouldAutoRefresh,
  }
}
