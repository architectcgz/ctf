<script setup lang="ts">
import { Play } from 'lucide-vue-next'

import type { AWDInstanceOrchestrationRowView } from './awdInstanceOrchestration.types'

defineProps<{
  row: AWDInstanceOrchestrationRowView
  loading: boolean
  hasPendingStart: boolean
}>()

const emit = defineEmits<{
  'start-cell': [teamId: string, serviceId: string]
  'start-team': [teamId: string]
}>()
</script>

<template>
  <tr>
    <th class="team-cell">
      <span class="team-name">{{ row.teamName }}</span>
      <span class="team-meta">Captain {{ row.captainId }}</span>
    </th>
    <td
      v-for="cell in row.cells"
      :key="`${row.teamId}:${cell.serviceId}`"
      class="service-cell"
    >
      <div class="instance-cell">
        <template v-if="cell.status">
          <span
            class="instance-status"
            :class="`instance-status--${cell.status}`"
          >
            {{ cell.statusLabel }}
          </span>
          <a
            v-if="cell.accessUrl"
            class="instance-link"
            :href="cell.accessUrl"
            target="_blank"
            rel="noreferrer"
          >
            访问
          </a>
        </template>
        <button
          v-else
          type="button"
          class="cell-start-btn"
          :disabled="loading || hasPendingStart"
          @click="emit('start-cell', row.teamId, cell.serviceId)"
        >
          <Play
            class="btn-icon"
            :class="{ 'animate-spin': cell.isStarting }"
          />
          <span>{{ cell.isStarting ? '启动中' : '启动' }}</span>
        </button>
      </div>
    </td>
    <td class="row-action-cell">
      <button
        type="button"
        class="row-start-btn"
        :disabled="loading || hasPendingStart || !row.hasMissingService"
        @click="emit('start-team', row.teamId)"
      >
        启动本队
      </button>
    </td>
  </tr>
</template>
