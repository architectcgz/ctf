<script setup lang="ts">
import type { AWDTeamServiceData } from '@/api/contracts'

import type {
  AWDServiceStatusFilterOptionView,
  AWDServiceStatusTeamOptionView,
} from './awdServiceStatusPanel.types'

defineProps<{
  teamCount: number
  teamOptions: AWDServiceStatusTeamOptionView[]
  sourceOptions: AWDServiceStatusFilterOptionView[]
  alertOptions: AWDServiceStatusFilterOptionView[]
  serviceTeamFilter: string
  serviceStatusFilter: 'all' | AWDTeamServiceData['service_status']
  serviceCheckSourceFilter: string
  serviceAlertReasonFilter: string
}>()

const emit = defineEmits<{
  updateServiceTeamFilter: [value: string]
  updateServiceStatusFilter: [value: 'all' | AWDTeamServiceData['service_status']]
  updateServiceCheckSourceFilter: [value: string]
  updateServiceAlertReasonFilter: [value: string]
  exportServices: []
}>()

function updateServiceStatusFilter(value: string): void {
  if (value !== 'all' && value !== 'up' && value !== 'down' && value !== 'compromised') {
    return
  }
  emit('updateServiceStatusFilter', value)
}
</script>

<template>
  <div class="matrix-toolbar">
    <div class="toolbar-left">
      <h3 class="viewer-title">
        服务运行矩阵
      </h3>
      <div class="filter-summary">
        显示 {{ teamCount }} 支队伍
      </div>
    </div>

    <div class="toolbar-right">
      <div class="matrix-filters">
        <select
          id="awd-service-filter-team"
          :value="serviceTeamFilter"
          class="matrix-select"
          @change="emit('updateServiceTeamFilter', ($event.target as HTMLSelectElement).value)"
        >
          <option value="">
            所有队伍
          </option>
          <option
            v-for="team in teamOptions"
            :key="team.teamId"
            :value="team.teamId"
          >
            {{ team.teamName }}
          </option>
        </select>
        <select
          id="awd-service-filter-status"
          :value="serviceStatusFilter"
          class="matrix-select"
          @change="updateServiceStatusFilter(($event.target as HTMLSelectElement).value)"
        >
          <option value="all">
            所有状态
          </option>
          <option value="up">
            在线 (UP)
          </option>
          <option value="down">
            离线 (DOWN)
          </option>
          <option value="compromised">
            失陷 (EXP)
          </option>
        </select>
        <select
          id="awd-service-filter-source"
          :value="serviceCheckSourceFilter"
          class="matrix-select"
          @change="emit('updateServiceCheckSourceFilter', ($event.target as HTMLSelectElement).value)"
        >
          <option value="">
            所有来源
          </option>
          <option
            v-for="source in sourceOptions"
            :key="source.value"
            :value="source.value"
          >
            {{ source.label }}
          </option>
        </select>
        <select
          id="awd-service-filter-alert"
          :value="serviceAlertReasonFilter"
          class="matrix-select"
          @change="emit('updateServiceAlertReasonFilter', ($event.target as HTMLSelectElement).value)"
        >
          <option value="">
            所有告警
          </option>
          <option
            v-for="alert in alertOptions"
            :key="alert.value"
            :value="alert.value"
          >
            {{ alert.label }}
          </option>
        </select>
      </div>
      <button
        id="awd-export-services"
        type="button"
        class="ops-btn ops-btn--neutral"
        @click="emit('exportServices')"
      >
        导出报告
      </button>
    </div>
  </div>
</template>
