import { computed, type Ref } from 'vue'

import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTeamServiceData,
} from '@/api/contracts'

interface UseAwdInspectorSummaryMetricsOptions {
  rounds: Ref<AWDRoundData[]>
  selectedRoundId: Ref<string | null>
  services: Ref<AWDTeamServiceData[]>
  attacks: Ref<AWDAttackLogData[]>
  summary: Ref<AWDRoundSummaryData | null>
  checking: Ref<boolean>
}

export function useAwdInspectorSummaryMetrics({
  rounds,
  selectedRoundId,
  services,
  attacks,
  summary,
  checking,
}: UseAwdInspectorSummaryMetricsOptions) {
  const selectedRound = computed(
    () => rounds.value.find((item) => item.id === selectedRoundId.value) || null
  )
  const summaryMetrics = computed(() => summary.value?.metrics || null)
  const totalServiceCount = computed(
    () => summaryMetrics.value?.total_service_count ?? services.value.length
  )
  const totalAttackCount = computed(() => summaryMetrics.value?.total_attack_count ?? attacks.value.length)
  const upCount = computed(
    () =>
      summaryMetrics.value?.service_up_count ??
      services.value.filter((item) => item.service_status === 'up').length
  )
  const compromisedCount = computed(
    () =>
      summaryMetrics.value?.service_compromised_count ??
      services.value.filter((item) => item.service_status === 'compromised').length
  )
  const downCount = computed(
    () =>
      summaryMetrics.value?.service_down_count ??
      services.value.filter((item) => item.service_status === 'down').length
  )
  const successfulAttackCount = computed(
    () =>
      summaryMetrics.value?.successful_attack_count ??
      attacks.value.filter((item) => item.is_success).length
  )
  const failedAttackCount = computed(
    () =>
      summaryMetrics.value?.failed_attack_count ??
      attacks.value.filter((item) => !item.is_success).length
  )
  const attackedServiceCount = computed(
    () =>
      summaryMetrics.value?.attacked_service_count ??
      services.value.filter((item) => item.attack_received > 0).length
  )
  const defenseSuccessCount = computed(
    () =>
      summaryMetrics.value?.defense_success_count ??
      services.value.filter(
        (item) => item.attack_received > 0 && item.service_status === 'up'
      ).length
  )
  const manualCheckCount = computed(() => {
    if (summaryMetrics.value) {
      return (
        summaryMetrics.value.manual_current_round_check_count +
        summaryMetrics.value.manual_selected_round_check_count +
        summaryMetrics.value.manual_service_check_count
      )
    }
    return services.value.filter((item) => {
      const source =
        typeof item.check_result.check_source === 'string' ? item.check_result.check_source : ''
      return (
        source === 'manual_current_round' ||
        source === 'manual_selected_round' ||
        source === 'manual_service_check'
      )
    }).length
  })

  const checkButtonLabel = computed(() => {
    if (checking.value) {
      return '执行巡检中...'
    }
    if (!selectedRound.value) {
      return '立即巡检所选轮次'
    }
    if (selectedRound.value.status === 'running') {
      return '立即巡检当前轮'
    }
    return `重跑第 ${selectedRound.value.round_number} 轮巡检`
  })

  return {
    selectedRound,
    summaryMetrics,
    totalServiceCount,
    totalAttackCount,
    upCount,
    compromisedCount,
    downCount,
    successfulAttackCount,
    failedAttackCount,
    attackedServiceCount,
    defenseSuccessCount,
    manualCheckCount,
    checkButtonLabel,
  }
}
