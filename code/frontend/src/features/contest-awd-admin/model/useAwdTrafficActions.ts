import type { Ref } from 'vue'

import type { AWDTrafficFilterState } from './awdAdminSupport'

interface UseAwdTrafficActionsOptions {
  selectedRoundId: Ref<string | null>
  applyTrafficFiltersPatch: (
    patch: Partial<
      Pick<
        AWDTrafficFilterState,
        | 'attacker_team_id'
        | 'victim_team_id'
        | 'service_id'
        | 'awd_challenge_id'
        | 'status_group'
        | 'path_keyword'
      >
    >
  ) => void
  setTrafficPageState: (page: number) => void
  resetTrafficFiltersState: () => void
  refreshTrafficEvents: (roundId?: string | null) => Promise<void>
}

export function useAwdTrafficActions(options: UseAwdTrafficActionsOptions) {
  const {
    selectedRoundId,
    applyTrafficFiltersPatch,
    setTrafficPageState,
    resetTrafficFiltersState,
    refreshTrafficEvents,
  } = options

  async function applyTrafficFilters(
    patch: Partial<
      Pick<
        AWDTrafficFilterState,
        | 'attacker_team_id'
        | 'victim_team_id'
        | 'service_id'
        | 'awd_challenge_id'
        | 'status_group'
        | 'path_keyword'
      >
    >
  ) {
    applyTrafficFiltersPatch(patch)
    await refreshTrafficEvents(selectedRoundId.value)
  }

  async function setTrafficPage(page: number) {
    setTrafficPageState(page)
    await refreshTrafficEvents(selectedRoundId.value)
  }

  async function resetTrafficFilters() {
    resetTrafficFiltersState()
    await refreshTrafficEvents(selectedRoundId.value)
  }

  return {
    applyTrafficFilters,
    setTrafficPage,
    resetTrafficFilters,
  }
}
