<script setup lang="ts">
import { computed } from 'vue'
import type { AWDTeamServiceData } from '@/api/contracts'
import type {
  AWDServiceStatusPanelEmits,
  AWDServiceStatusPanelProps,
} from './awdInspector.types'
import AWDServiceRoundPerformanceTable from './AWDServiceRoundPerformanceTable.vue'
import AWDServiceStatusMatrix from './AWDServiceStatusMatrix.vue'
import AWDServiceStatusToolbar from './AWDServiceStatusToolbar.vue'
import './awdServiceStatusPanel.css'
import type {
  AWDServiceStatusChallengeColumnView,
  AWDServiceStatusFilterOptionView,
  AWDServiceStatusRowView,
  AWDServiceStatusTeamOptionView,
} from './awdServiceStatusPanel.types'

const props = defineProps<AWDServiceStatusPanelProps>()
const emit = defineEmits<AWDServiceStatusPanelEmits>()

// Matrix specific derivations
const distinctChallengeIds = computed(() => {
  const ids = new Set<string>()
  props.services.forEach(s => ids.add(s.awd_challenge_id))
  return Array.from(ids)
})

const teamMap = computed(() => {
  const map = new Map<
    string,
    {
      team_name: string
      services: Record<string, AWDTeamServiceData>
    }
  >()

  props.filteredServices.forEach(s => {
    if (!map.has(s.team_id)) {
      map.set(s.team_id, { team_name: s.team_name, services: {} })
    }
    map.get(s.team_id)!.services[s.awd_challenge_id] = s
  })
  return Array.from(map.entries()).sort((a, b) => a[1].team_name.localeCompare(b[1].team_name))
})

function getChallengeLabel(challengeId: string): string {
  return props.getChallengeTitle(challengeId) || `题目 ${challengeId}`
}

function getServiceCellKey(teamId: string, challengeId: string): string {
  return `${teamId}-${challengeId}`
}

function getServicePresentationResult(service: AWDTeamServiceData): Record<string, unknown> {
  return props.getServiceCheckPresentationResult(service)
}

function getServiceCheckerLabel(service: AWDTeamServiceData): string {
  const result = getServicePresentationResult(service)
  return props.getCheckerTypeLabel(result.checker_type || service.checker_type) || '未标注'
}

function getServiceSourceLabel(service: AWDTeamServiceData): string {
  const result = getServicePresentationResult(service)
  return props.getCheckSourceLabel(result.check_source) || '未标注'
}

function getServiceStatusReasonLabel(service: AWDTeamServiceData): string {
  const result = getServicePresentationResult(service)
  const previewPassCount =
    typeof result.preview_pass_count === 'number' ? result.preview_pass_count : undefined
  const previewTotalCount =
    typeof result.preview_total_count === 'number' ? result.preview_total_count : undefined

  if (
    typeof previewPassCount === 'number' &&
    typeof previewTotalCount === 'number' &&
    Number.isFinite(previewPassCount) &&
    Number.isFinite(previewTotalCount) &&
    previewTotalCount > 0
  ) {
    return `${previewPassCount}/${previewTotalCount} 通过`
  }

  return props.getCheckStatusLabel(result.status_reason) || '未返回'
}

function getServiceCheckedAtLabel(service: AWDTeamServiceData): string {
  const result = getServicePresentationResult(service)
  const checkedAt =
    typeof result.checked_at === 'string' && result.checked_at.trim() !== ''
      ? result.checked_at
      : service.updated_at
  return props.formatDateTime(checkedAt)
}

const teamOptions = computed<AWDServiceStatusTeamOptionView[]>(() =>
  props.serviceTeamOptions.map((team) => ({
    teamId: team.team_id,
    teamName: team.team_name,
  }))
)

const sourceOptions = computed<AWDServiceStatusFilterOptionView[]>(() =>
  props.serviceCheckSourceOptions.map((source) => ({
    value: source,
    label: props.getCheckSourceLabel(source),
  }))
)

const alertOptions = computed<AWDServiceStatusFilterOptionView[]>(() =>
  props.serviceAlerts.map((alert) => ({
    value: alert.key,
    label: alert.label,
  }))
)

const challengeColumns = computed<AWDServiceStatusChallengeColumnView[]>(() =>
  distinctChallengeIds.value.map((challengeId) => ({
    challengeId,
    label: getChallengeLabel(challengeId),
  }))
)

const serviceStatusRows = computed<AWDServiceStatusRowView[]>(() =>
  teamMap.value.map(([teamId, team]) => ({
    teamId,
    teamName: team.team_name,
    cells: challengeColumns.value.map((column) => {
      const service = team.services[column.challengeId]
      if (!service) {
        return {
          key: getServiceCellKey(teamId, column.challengeId),
          status: null,
          statusClass: '',
          statusLabel: '',
          checkerLabel: '',
          sourceLabel: '',
          reasonLabel: '',
          checkedAtLabel: '',
        }
      }

      return {
        key: getServiceCellKey(teamId, column.challengeId),
        status: service.service_status,
        statusClass: props.getServiceStatusClass(service.service_status),
        statusLabel: props.getServiceStatusLabel(service.service_status),
        checkerLabel: getServiceCheckerLabel(service),
        sourceLabel: getServiceSourceLabel(service),
        reasonLabel: getServiceStatusReasonLabel(service),
        checkedAtLabel: getServiceCheckedAtLabel(service),
      }
    }),
  }))
)

const emptyMessage = computed(() =>
  props.services.length === 0 ? '当前轮次还没有服务巡检记录' : '没有找到匹配的服务项'
)
</script>

<template>
  <div class="awd-matrix-viewer">
    <AWDServiceStatusToolbar
      :team-count="teamMap.length"
      :team-options="teamOptions"
      :source-options="sourceOptions"
      :alert-options="alertOptions"
      :service-team-filter="serviceTeamFilter"
      :service-status-filter="serviceStatusFilter"
      :service-check-source-filter="serviceCheckSourceFilter"
      :service-alert-reason-filter="serviceAlertReasonFilter"
      @update-service-team-filter="emit('updateServiceTeamFilter', $event)"
      @update-service-status-filter="emit('updateServiceStatusFilter', $event)"
      @update-service-check-source-filter="emit('updateServiceCheckSourceFilter', $event)"
      @update-service-alert-reason-filter="emit('updateServiceAlertReasonFilter', $event)"
      @export-services="emit('exportServices')"
    />

    <AWDServiceStatusMatrix
      :columns="challengeColumns"
      :rows="serviceStatusRows"
      :empty-message="emptyMessage"
    />

    <AWDServiceRoundPerformanceTable
      v-if="summary"
      :summary="summary"
    />
  </div>
</template>
