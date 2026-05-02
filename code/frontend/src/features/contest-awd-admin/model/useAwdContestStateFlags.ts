import { computed, type Ref } from 'vue'

import type { AWDRoundData, ContestDetailData } from '@/api/contracts'

interface UseAwdContestStateFlagsOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
  selectedRound: Readonly<Ref<AWDRoundData | null>>
}

export function useAwdContestStateFlags(options: UseAwdContestStateFlagsOptions) {
  const { selectedContest, selectedRound } = options

  const hasSelectedContest = computed(
    () => Boolean(selectedContest.value) && selectedContest.value?.mode === 'awd'
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
    shouldAutoRefresh,
  }
}
