<script setup lang="ts">
import { ref, toRef, computed } from 'vue'
import type { AWDTeamServiceData } from '@/api/contracts'
import AWDInspectorCanvasWorkspace from './AWDInspectorCanvasWorkspace.vue'
import AWDInspectorStatsHud from './AWDInspectorStatsHud.vue'
import AWDRoundHeaderPanel from './AWDRoundHeaderPanel.vue'
import type {
  AWDAttackLogPanelProps,
  AWDInspectorSubTab,
  AWDRoundInspectorEmits,
  AWDRoundInspectorProps,
  AWDScoreboardSummaryPanelProps,
  AWDServiceAlertView,
  AWDServiceStatusPanelProps,
  AWDTrafficPanelProps,
} from './awdInspector.types'
import './awdRoundInspector.css'
import {
  useAwdCheckResultPresentation,
  useAwdInspectorCoreState,
  useAwdInspectorDerivedData,
  useAwdInspectorExports,
  useAwdInspectorFormatting,
} from '@/features/awd-inspector/model'

const props = defineProps<AWDRoundInspectorProps & { initialTab?: 'matrix' | 'attacks' | 'traffic' | 'scoreboard', hideStudioLink?: boolean }>()

const emit = defineEmits<AWDRoundInspectorEmits & { 'open:contestEdit': [] }>()
defineSlots<{
  'service-alerts'?: (props: {
    serviceAlerts: AWDServiceAlertView[]
    selectedAlertKey: string
    getServiceAlertClass: (alertKey: string) => string
    applyServiceAlertFilter: (alertKey: string) => void
  }) => unknown
}>()

const activeSubTab = ref<AWDInspectorSubTab>(props.initialTab || 'matrix')

const {
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
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
  manualCheckCount,
  checkButtonLabel,
} = useAwdInspectorCoreState({
  contest: toRef(props, 'contest'),
  selectedRoundId: toRef(props, 'selectedRoundId'),
  rounds: toRef(props, 'rounds'),
  services: toRef(props, 'services'),
  attacks: toRef(props, 'attacks'),
  summary: toRef(props, 'summary'),
  checking: toRef(props, 'checking'),
})

const {
  formatDateTime,
  getRoundStatusLabel,
  getRoundStatusClass,
  getServiceStatusLabel,
  getServiceStatusClass,
  getAttackTypeLabel,
  getAttackSourceLabel,
  formatPercent,
  getTrafficStatusGroupLabel,
  getTrafficStatusGroupClass,
  getChallengeTitle,
  buildExportFilename,
  getSelectedRoundLabel,
  formatScore,
  getSourceOverviewLabel,
  getSourceOverviewDescription,
} = useAwdInspectorFormatting({
  contest: toRef(props, 'contest'),
  challengeLinks: toRef(props, 'challengeLinks'),
  selectedRound,
  summaryMetrics,
  manualCheckCount,
})

const {
  getCheckSourceLabel,
  getCheckerTypeLabel,
  getCheckStatusLabel,
  summarizeCheckResult,
  getCheckActions,
  getCheckTargets,
  getTargetActions,
  getTargetProbeSummary,
  getProbeStatusText,
  formatLatency,
} = useAwdCheckResultPresentation({
  formatDateTime,
})

const {
  getServiceCheckSourceValue,
  serviceTeamOptions,
  serviceCheckSourceOptions,
  serviceAlerts,
  filteredServices,
  attackTeamOptions,
  attackSourceOptions,
  trafficTeamOptions,
  filteredAttacks,
  getServiceAlertClass,
  applyServiceAlertFilter,
} = useAwdInspectorDerivedData({
  services: toRef(props, 'services'),
  attacks: toRef(props, 'attacks'),
  trafficSummary: toRef(props, 'trafficSummary'),
  trafficEvents: toRef(props, 'trafficEvents'),
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
  getChallengeTitle,
  getCheckStatusLabel,
})

const {
  getTrafficTeamName,
  getTrafficChallengeTitle,
  getTrafficSourceLabel,
  exportFilteredServices,
  exportFilteredAttacks,
  exportReviewPackage,
} = useAwdInspectorExports({
  contest: toRef(props, 'contest'),
  selectedRound,
  summary: toRef(props, 'summary'),
  scoreboardRows: toRef(props, 'scoreboardRows'),
  scoreboardFrozen: toRef(props, 'scoreboardFrozen'),
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
  serviceTeamOptions,
  attackTeamOptions,
  trafficTeamOptions,
  serviceAlerts,
  filteredServices,
  filteredAttacks,
  formatDateTime,
  getChallengeTitle,
  getSelectedRoundLabel,
  buildExportFilename,
  getServiceStatusLabel,
  getAttackTypeLabel,
  getAttackSourceLabel,
  getCheckSourceLabel,
  getCheckerTypeLabel,
  getServiceAlertLabel: (s: any) => s,
  summarizeCheckResult,
  getServiceCheckSourceValue,
})

function getServiceCheckPresentationResult(service: AWDTeamServiceData): Record<string, unknown> {
  return {
    checker_type: service.checker_type,
    ...service.check_result,
  }
}

const serviceStatusPanelProps = computed<AWDServiceStatusPanelProps>(() => ({
  services: props.services,
  filteredServices: filteredServices.value,
  summary: props.summary,
  serviceAlerts: serviceAlerts.value,
  serviceTeamOptions: serviceTeamOptions.value,
  serviceCheckSourceOptions: serviceCheckSourceOptions.value,
  serviceTeamFilter: serviceTeamFilter.value,
  serviceStatusFilter: serviceStatusFilter.value,
  serviceCheckSourceFilter: serviceCheckSourceFilter.value,
  serviceAlertReasonFilter: serviceAlertReasonFilter.value,
  getChallengeTitle,
  getServiceStatusLabel,
  getServiceStatusClass,
  getCheckSourceLabel,
  getCheckerTypeLabel,
  getCheckStatusLabel,
  summarizeCheckResult,
  getCheckActions,
  getCheckTargets,
  getTargetActions,
  getTargetProbeSummary,
  getProbeStatusText,
  formatDateTime,
  formatLatency,
  getServiceCheckPresentationResult,
}))

const scoreboardSummaryPanelProps = computed<AWDScoreboardSummaryPanelProps>(() => ({
  scoreboardRows: props.scoreboardRows,
  scoreboardFrozen: props.scoreboardFrozen,
  summary: props.summary,
  formatScore,
  formatDateTime,
}))

const attackLogPanelProps = computed<AWDAttackLogPanelProps>(() => ({
  attacks: props.attacks,
  filteredAttacks: filteredAttacks.value,
  attackTeamOptions: attackTeamOptions.value,
  attackSourceOptions: attackSourceOptions.value,
  attackTeamFilter: attackTeamFilter.value,
  attackResultFilter: attackResultFilter.value,
  attackSourceFilter: attackSourceFilter.value,
  formatDateTime,
  getChallengeTitle,
  getAttackTypeLabel,
  getAttackSourceLabel,
}))

const trafficPanelProps = computed<AWDTrafficPanelProps>(() => ({
  updatedAt: selectedRound.value?.updated_at,
  challengeLinks: props.challengeLinks,
  trafficSummary: props.trafficSummary,
  trafficEvents: props.trafficEvents,
  trafficEventsTotal: props.trafficEventsTotal,
  trafficFilters: props.trafficFilters,
  trafficTeamOptions: trafficTeamOptions.value,
  loadingTrafficSummary: props.loadingTrafficSummary,
  loadingTrafficEvents: props.loadingTrafficEvents,
  formatDateTime,
  formatPercent,
  getTrafficStatusGroupLabel,
  getTrafficStatusGroupClass,
  getTrafficTeamName,
  getTrafficChallengeTitle,
  getTrafficSourceLabel,
}))
</script>

<template>
  <div class="awd-inspector-workbench">
    <AWDRoundHeaderPanel
      :contest="contest"
      :rounds="rounds"
      :rounds-count="rounds.length"
      :selected-round="selectedRound"
      :selected-round-id="selectedRoundId"
      :loading-rounds="loadingRounds"
      :loading-round-detail="loadingRoundDetail"
      :checking="checking"
      :should-auto-refresh="shouldAutoRefresh"
      :can-record-service-checks="canRecordServiceChecks"
      :can-record-attack-logs="canRecordAttackLogs"
      :service-check-hint="serviceCheckHint"
      :attack-log-hint="attackLogHint"
      :compromised-count="compromisedCount"
      :total-attack-count="totalAttackCount"
      :successful-attack-count="successfulAttackCount"
      :failed-attack-count="failedAttackCount"
      :get-round-status-label="getRoundStatusLabel"
      :get-round-status-class="getRoundStatusClass"
      :check-button-label="checkButtonLabel"
      :hide-studio-link="hideStudioLink"
      @refresh="emit('refresh')"
      @open-create-round-dialog="emit('openCreateRoundDialog')"
      @open-service-check-dialog="emit('openServiceCheckDialog')"
      @open-attack-log-dialog="emit('openAttackLogDialog')"
      @run-selected-round-check="emit('runSelectedRoundCheck')"
      @update:selected-round-id="emit('update:selectedRoundId', $event)"
      @open:contest-edit="emit('open:contestEdit')"
    />

    <AWDInspectorStatsHud
      v-if="selectedRound"
      :total-service-count="totalServiceCount"
      :up-count="upCount"
      :down-count="downCount"
      :total-attack-count="totalAttackCount"
      :successful-attack-count="successfulAttackCount"
      :compromised-count="compromisedCount"
      :attacked-service-count="attackedServiceCount"
      :get-source-overview-label="getSourceOverviewLabel"
      :get-source-overview-description="getSourceOverviewDescription"
    />

    <AWDInspectorCanvasWorkspace
      :active-sub-tab="activeSubTab"
      :selected-round="selectedRound"
      :loading-round-detail="loadingRoundDetail"
      :service-alerts="serviceAlerts"
      :selected-alert-key="serviceAlertReasonFilter"
      :get-service-alert-class="getServiceAlertClass"
      :apply-service-alert-filter="applyServiceAlertFilter"
      :service-status-panel="serviceStatusPanelProps"
      :scoreboard-summary-panel="scoreboardSummaryPanelProps"
      :attack-log-panel="attackLogPanelProps"
      :traffic-panel="trafficPanelProps"
      @update:active-sub-tab="activeSubTab = $event"
      @export-review-package="exportReviewPackage"
      @update-service-team-filter="serviceTeamFilter = $event"
      @update-service-status-filter="serviceStatusFilter = $event"
      @update-service-check-source-filter="serviceCheckSourceFilter = $event"
      @update-service-alert-reason-filter="serviceAlertReasonFilter = $event"
      @export-services="exportFilteredServices"
      @update-attack-team-filter="attackTeamFilter = $event"
      @update-attack-result-filter="attackResultFilter = $event"
      @update-attack-source-filter="attackSourceFilter = $event"
      @export-attacks="exportFilteredAttacks"
      @apply-traffic-filters="emit('applyTrafficFilters', $event)"
      @change-traffic-page="emit('changeTrafficPage', $event)"
      @reset-traffic-filters="emit('resetTrafficFilters')"
    >
      <template #service-alerts="slotProps">
        <slot
          name="service-alerts"
          :service-alerts="slotProps.serviceAlerts"
          :selected-alert-key="slotProps.selectedAlertKey"
          :get-service-alert-class="slotProps.getServiceAlertClass"
          :apply-service-alert-filter="slotProps.applyServiceAlertFilter"
        />
      </template>
    </AWDInspectorCanvasWorkspace>
  </div>
</template>
