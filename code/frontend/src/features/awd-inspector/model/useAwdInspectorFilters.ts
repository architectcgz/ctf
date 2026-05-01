import { ref, watch, type Ref } from 'vue'

import type { AWDAttackLogData, AWDTeamServiceData } from '@/api/contracts'

interface AwdInspectorFilterState {
  service_team_id: string
  service_status: 'all' | AWDTeamServiceData['service_status']
  service_check_source: string
  service_alert_reason: string
  attack_team_id: string
  attack_result: 'all' | 'success' | 'failed'
  attack_source: 'all' | AWDAttackLogData['source']
}

interface UseAwdInspectorFiltersOptions {
  contestId: Ref<string>
  selectedRoundId: Ref<string | null>
}

function getStorageKey(contestId: string, roundId: string): string {
  return `ctf_admin_awd_filters:${contestId}:${roundId}`
}

function loadFilterState(contestId: string, roundId: string): AwdInspectorFilterState | null {
  if (typeof window === 'undefined') return null

  const raw = window.sessionStorage.getItem(getStorageKey(contestId, roundId))
  if (!raw) return null

  try {
    const parsed = JSON.parse(raw) as Partial<AwdInspectorFilterState>
    return {
      service_team_id: typeof parsed.service_team_id === 'string' ? parsed.service_team_id : '',
      service_status:
        parsed.service_status === 'up' ||
        parsed.service_status === 'down' ||
        parsed.service_status === 'compromised'
          ? parsed.service_status
          : 'all',
      service_check_source:
        typeof parsed.service_check_source === 'string' ? parsed.service_check_source : '',
      service_alert_reason:
        typeof parsed.service_alert_reason === 'string' ? parsed.service_alert_reason : '',
      attack_team_id: typeof parsed.attack_team_id === 'string' ? parsed.attack_team_id : '',
      attack_result:
        parsed.attack_result === 'success' || parsed.attack_result === 'failed'
          ? parsed.attack_result
          : 'all',
      attack_source:
        parsed.attack_source === 'manual_attack_log' ||
        parsed.attack_source === 'submission' ||
        parsed.attack_source === 'legacy'
          ? parsed.attack_source
          : 'all',
    }
  } catch {
    return null
  }
}

function persistFilterState(
  contestId: string,
  roundId: string,
  state: AwdInspectorFilterState
): void {
  if (typeof window === 'undefined') return
  window.sessionStorage.setItem(getStorageKey(contestId, roundId), JSON.stringify(state))
}

export function useAwdInspectorFilters({
  contestId,
  selectedRoundId,
}: UseAwdInspectorFiltersOptions) {
  const serviceTeamFilter = ref('')
  const serviceStatusFilter = ref<'all' | AWDTeamServiceData['service_status']>('all')
  const serviceCheckSourceFilter = ref('')
  const serviceAlertReasonFilter = ref('')
  const attackTeamFilter = ref('')
  const attackResultFilter = ref<'all' | 'success' | 'failed'>('all')
  const attackSourceFilter = ref<'all' | AWDAttackLogData['source']>('all')

  let syncingPersistedFilters = false

  function resetFilters(): void {
    serviceTeamFilter.value = ''
    serviceStatusFilter.value = 'all'
    serviceCheckSourceFilter.value = ''
    serviceAlertReasonFilter.value = ''
    attackTeamFilter.value = ''
    attackResultFilter.value = 'all'
    attackSourceFilter.value = 'all'
  }

  watch(
    [contestId, selectedRoundId],
    ([nextContestId, nextRoundId]) => {
      syncingPersistedFilters = true
      if (!nextRoundId) {
        resetFilters()
        syncingPersistedFilters = false
        return
      }

      const storedState = loadFilterState(nextContestId, nextRoundId)
      if (!storedState) {
        resetFilters()
        syncingPersistedFilters = false
        return
      }

      serviceTeamFilter.value = storedState.service_team_id
      serviceStatusFilter.value = storedState.service_status
      serviceCheckSourceFilter.value = storedState.service_check_source
      serviceAlertReasonFilter.value = storedState.service_alert_reason
      attackTeamFilter.value = storedState.attack_team_id
      attackResultFilter.value = storedState.attack_result
      attackSourceFilter.value = storedState.attack_source
      syncingPersistedFilters = false
    },
    { immediate: true }
  )

  watch(
    [
      contestId,
      selectedRoundId,
      serviceTeamFilter,
      serviceStatusFilter,
      serviceCheckSourceFilter,
      serviceAlertReasonFilter,
      attackTeamFilter,
      attackResultFilter,
      attackSourceFilter,
    ],
    ([nextContestId, nextRoundId]) => {
      if (syncingPersistedFilters || !nextRoundId) return

      persistFilterState(nextContestId, nextRoundId, {
        service_team_id: serviceTeamFilter.value,
        service_status: serviceStatusFilter.value,
        service_check_source: serviceCheckSourceFilter.value,
        service_alert_reason: serviceAlertReasonFilter.value,
        attack_team_id: attackTeamFilter.value,
        attack_result: attackResultFilter.value,
        attack_source: attackSourceFilter.value,
      })
    },
    { immediate: true }
  )

  return {
    serviceTeamFilter,
    serviceStatusFilter,
    serviceCheckSourceFilter,
    serviceAlertReasonFilter,
    attackTeamFilter,
    attackResultFilter,
    attackSourceFilter,
    resetFilters,
  }
}
