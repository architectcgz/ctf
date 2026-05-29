<script setup lang="ts">
import { computed, toRef } from 'vue'

import type { AWDTrafficStatusGroup } from '@/api/contracts'
import type {
  AWDTrafficPanelEmits,
  AWDTrafficPanelProps,
} from './awdInspector.types'
import AWDTrafficEventTable from './AWDTrafficEventTable.vue'
import AWDTrafficIntelligenceGrid from './AWDTrafficIntelligenceGrid.vue'
import AWDTrafficSummaryBand from './AWDTrafficSummaryBand.vue'
import './awdTrafficPanel.css'
import type { AWDTrafficServiceOptionView } from './awdTrafficPanel.types'
import { useAwdTrafficPanel } from '@/features/awd-inspector/model'

const props = defineProps<AWDTrafficPanelProps>()
const emit = defineEmits<AWDTrafficPanelEmits>()

const serviceOptions = computed<AWDTrafficServiceOptionView[]>(() => {
  const seen = new Set<string>()
  return props.challengeLinks
    .filter((item) => {
      const serviceId = item.awd_service_id?.trim()
      if (!serviceId || seen.has(serviceId)) {
        return false
      }
      seen.add(serviceId)
      return true
    })
    .map((item) => ({
      serviceId: item.awd_service_id!,
      title: item.title || `Challenge #${item.challenge_id}`,
    }))
})

const {
  trafficPathKeywordInput,
  trafficTotalPages,
  trafficTrendRows,
  trafficSummaryStats,
  trafficTrendNarrative,
  trafficStatusGroupOptions,
  applyTrafficKeywordFilter,
  onTrafficStatusGroupChange,
  applyTrafficFilterPatch,
  handleTrafficPageChange,
} = useAwdTrafficPanel({
  trafficSummary: toRef(props, 'trafficSummary'),
  trafficEventsTotal: toRef(props, 'trafficEventsTotal'),
  trafficFilters: toRef(props, 'trafficFilters'),
  loadingTrafficEvents: toRef(props, 'loadingTrafficEvents'),
  trafficPathKeyword: computed(() => props.trafficFilters.path_keyword),
  formatDateTime: props.formatDateTime,
  formatPercent: props.formatPercent,
  applyTrafficFilters: (patch) => emit('applyTrafficFilters', patch),
  changeTrafficPage: (page) => emit('changeTrafficPage', page),
})

function handleTrafficStatusGroupChange(value: string): void {
  onTrafficStatusGroupChange(value)
}

function getTrafficStatusGroupLabel(statusGroup: AWDTrafficStatusGroup): string {
  return props.getTrafficStatusGroupLabel(statusGroup)
}

const hasTrafficSummary = computed(() => Boolean(props.trafficSummary))

function updateTrafficPathKeywordInput(value: string): void {
  trafficPathKeywordInput.value = value
}
</script>

<template>
  <div class="studio-traffic-analysis">
    <AWDTrafficSummaryBand
      v-if="hasTrafficSummary"
      :stats="trafficSummaryStats"
    />

    <AWDTrafficIntelligenceGrid
      v-if="trafficSummary"
      :summary="trafficSummary"
      :trend-rows="trafficTrendRows"
      :trend-narrative="trafficTrendNarrative"
    />

    <AWDTrafficEventTable
      :updated-at="updatedAt"
      :service-options="serviceOptions"
      :traffic-events="trafficEvents"
      :traffic-events-total="trafficEventsTotal"
      :traffic-filters="trafficFilters"
      :traffic-team-options="trafficTeamOptions"
      :traffic-status-group-options="trafficStatusGroupOptions"
      :traffic-path-keyword-input="trafficPathKeywordInput"
      :traffic-total-pages="trafficTotalPages"
      :loading-traffic-events="loadingTrafficEvents"
      :format-date-time="formatDateTime"
      :get-traffic-team-name="getTrafficTeamName"
      :get-traffic-challenge-title="getTrafficChallengeTitle"
      :get-traffic-status-group-label="getTrafficStatusGroupLabel"
      :get-traffic-status-group-class="getTrafficStatusGroupClass"
      @update-traffic-path-keyword-input="updateTrafficPathKeywordInput"
      @apply-traffic-keyword-filter="applyTrafficKeywordFilter"
      @handle-traffic-status-group-change="handleTrafficStatusGroupChange"
      @apply-traffic-filter-patch="applyTrafficFilterPatch"
      @reset-traffic-filters="emit('resetTrafficFilters')"
      @handle-traffic-page-change="handleTrafficPageChange"
    />
  </div>
</template>
