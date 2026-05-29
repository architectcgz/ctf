<script setup lang="ts">
import { Server } from 'lucide-vue-next'

import type {
  AWDInstanceOrchestrationRowView,
  AWDInstanceOrchestrationServiceView,
} from './awdInstanceOrchestration.types'
import AWDInstanceOrchestrationRow from './AWDInstanceOrchestrationRow.vue'

defineProps<{
  services: AWDInstanceOrchestrationServiceView[]
  rows: AWDInstanceOrchestrationRowView[]
  loading: boolean
  hasPendingStart: boolean
}>()

const emit = defineEmits<{
  'start-cell': [teamId: string, serviceId: string]
  'start-team': [teamId: string]
}>()
</script>

<template>
  <div
    v-if="rows.length === 0 || services.length === 0"
    class="orchestration-empty"
  >
    <Server class="empty-icon" />
    <span>暂无可编排的队伍或可见服务</span>
  </div>

  <div
    v-else
    class="orchestration-table-wrap"
  >
    <table class="orchestration-table">
      <colgroup>
        <col class="team-col-track">
        <col
          v-for="service in services"
          :key="`service-col-${service.serviceId}`"
          class="service-col-track"
        >
        <col class="action-col-track">
      </colgroup>
      <thead>
        <tr>
          <th class="team-col">
            队伍
          </th>
          <th
            v-for="service in services"
            :key="service.serviceId"
            class="service-col"
          >
            {{ service.displayName }}
          </th>
          <th class="action-col">
            操作
          </th>
        </tr>
      </thead>
      <tbody>
        <AWDInstanceOrchestrationRow
          v-for="row in rows"
          :key="row.teamId"
          :row="row"
          :loading="loading"
          :has-pending-start="hasPendingStart"
          @start-cell="(teamId, serviceId) => emit('start-cell', teamId, serviceId)"
          @start-team="(teamId) => emit('start-team', teamId)"
        />
      </tbody>
    </table>
  </div>
</template>
