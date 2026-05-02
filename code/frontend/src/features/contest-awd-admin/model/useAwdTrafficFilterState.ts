import { ref } from 'vue'

import {
  createDefaultTrafficFilters,
  type AWDTrafficFilterState,
} from './awdAdminSupport'

type AWDTrafficFilterPatch = Partial<
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

export function useAwdTrafficFilterState() {
  const trafficFilters = ref<AWDTrafficFilterState>(createDefaultTrafficFilters())

  function buildTrafficEventsParams() {
    const filters = trafficFilters.value
    return {
      attacker_team_id: filters.attacker_team_id || undefined,
      victim_team_id: filters.victim_team_id || undefined,
      service_id: filters.service_id || undefined,
      awd_challenge_id: filters.awd_challenge_id || undefined,
      status_group: filters.status_group === 'all' ? undefined : filters.status_group,
      path_keyword: filters.path_keyword.trim() || undefined,
      page: filters.page,
      page_size: filters.page_size,
    }
  }

  function applyTrafficFiltersPatch(patch: AWDTrafficFilterPatch) {
    trafficFilters.value = {
      ...trafficFilters.value,
      ...patch,
      page: 1,
    }
  }

  function setTrafficPageState(page: number) {
    const normalizedPage = Number.isFinite(page) && page > 0 ? Math.floor(page) : 1
    trafficFilters.value = {
      ...trafficFilters.value,
      page: normalizedPage,
    }
  }

  function syncTrafficPagination(page: number, pageSize: number) {
    trafficFilters.value = {
      ...trafficFilters.value,
      page,
      page_size: pageSize,
    }
  }

  function resetTrafficFiltersState() {
    trafficFilters.value = createDefaultTrafficFilters()
  }

  return {
    trafficFilters,
    buildTrafficEventsParams,
    applyTrafficFiltersPatch,
    setTrafficPageState,
    syncTrafficPagination,
    resetTrafficFiltersState,
  }
}
