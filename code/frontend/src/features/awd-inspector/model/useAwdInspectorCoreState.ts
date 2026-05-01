import { computed, type Ref } from 'vue'

import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTeamServiceData,
  ContestDetailData,
} from '@/api/contracts'
import { useAwdInspectorFilters } from './useAwdInspectorFilters'
import { useAwdInspectorSummaryMetrics } from './useAwdInspectorSummaryMetrics'

interface UseAwdInspectorCoreStateOptions {
  contest: Ref<ContestDetailData>
  selectedRoundId: Ref<string | null>
  rounds: Ref<AWDRoundData[]>
  services: Ref<AWDTeamServiceData[]>
  attacks: Ref<AWDAttackLogData[]>
  summary: Ref<AWDRoundSummaryData | null>
  checking: Ref<boolean>
}

export function useAwdInspectorCoreState({
  contest,
  selectedRoundId,
  rounds,
  services,
  attacks,
  summary,
  checking,
}: UseAwdInspectorCoreStateOptions) {
  const contestId = computed(() => contest.value.id)
  const filters = useAwdInspectorFilters({
    contestId,
    selectedRoundId,
  })

  const metrics = useAwdInspectorSummaryMetrics({
    rounds,
    selectedRoundId,
    services,
    attacks,
    summary,
    checking,
  })

  return {
    ...filters,
    ...metrics,
  }
}
