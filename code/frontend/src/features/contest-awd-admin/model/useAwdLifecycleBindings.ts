import { onBeforeUnmount, watch, type ComputedRef, type Ref } from 'vue'

import type { ContestDetailData } from '@/api/contracts'

import { persistSelectedRoundId } from './awdAdminSupport'

const AWD_AUTO_REFRESH_INTERVAL_MS = 15_000

interface UseAwdLifecycleBindingsOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
  selectedRoundId: Ref<string | null>
  shouldAutoRefresh: ComputedRef<boolean>
  refresh: () => Promise<void>
  refreshRoundDetail: (roundId: string) => Promise<void>
  clearRoundDetail: () => void
  resetTrafficFiltersState: () => void
  isSyncingSelectedRound: () => boolean
}

export function useAwdLifecycleBindings(options: UseAwdLifecycleBindingsOptions) {
  const {
    selectedContest,
    selectedRoundId,
    shouldAutoRefresh,
    refresh,
    refreshRoundDetail,
    clearRoundDetail,
    resetTrafficFiltersState,
    isSyncingSelectedRound,
  } = options

  let autoRefreshTimer: number | null = null

  function stopAutoRefresh() {
    if (autoRefreshTimer !== null) {
      window.clearInterval(autoRefreshTimer)
      autoRefreshTimer = null
    }
  }

  watch(
    () => selectedContest.value?.id || null,
    async (nextContestId, previousContestId) => {
      if (nextContestId !== previousContestId) {
        resetTrafficFiltersState()
      }
      await refresh()
    },
    { immediate: true }
  )

  watch(
    () => selectedRoundId.value,
    async (nextRoundId, previousRoundId) => {
      if (!selectedContest.value || !nextRoundId || nextRoundId === previousRoundId) {
        if (!nextRoundId) {
          clearRoundDetail()
        }
        return
      }
      if (isSyncingSelectedRound()) {
        return
      }
      await refreshRoundDetail(nextRoundId)
    }
  )

  watch(
    () => [selectedContest.value?.id || null, selectedRoundId.value] as const,
    ([contestId, roundId]) => {
      if (!contestId) {
        return
      }
      persistSelectedRoundId(contestId, roundId)
    },
    { immediate: true }
  )

  watch(
    shouldAutoRefresh,
    (enabled) => {
      stopAutoRefresh()
      if (!enabled || typeof window === 'undefined') {
        return
      }
      autoRefreshTimer = window.setInterval(() => {
        void refresh()
      }, AWD_AUTO_REFRESH_INTERVAL_MS)
    },
    { immediate: true }
  )

  onBeforeUnmount(() => {
    stopAutoRefresh()
  })
}
