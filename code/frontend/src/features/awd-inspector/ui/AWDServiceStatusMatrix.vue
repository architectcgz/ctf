<script setup lang="ts">
import { AlertCircle, CheckCircle2, SearchCheck, ShieldX } from 'lucide-vue-next'

import type {
  AWDServiceStatusChallengeColumnView,
  AWDServiceStatusRowView,
} from './awdServiceStatusPanel.types'

defineProps<{
  columns: AWDServiceStatusChallengeColumnView[]
  rows: AWDServiceStatusRowView[]
  emptyMessage: string
}>()

function getStatusIcon(status: AWDServiceStatusRowView['cells'][number]['status']) {
  if (status === 'up') {
    return CheckCircle2
  }
  if (status === 'compromised') {
    return ShieldX
  }
  return AlertCircle
}
</script>

<template>
  <div class="matrix-scroll custom-scrollbar">
    <table class="matrix-table">
      <thead>
        <tr>
          <th class="sticky-col header-team">
            队伍节点
          </th>
          <th
            v-for="column in columns"
            :key="column.challengeId"
          >
            {{ column.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="row in rows"
          :key="row.teamId"
        >
          <td class="sticky-col cell-team">
            <div class="team-name">
              {{ row.teamName }}
            </div>
          </td>
          <td
            v-for="cell in row.cells"
            :key="cell.key"
          >
            <div
              v-if="cell.status"
              class="status-box"
              :class="cell.statusClass"
            >
              <div
                class="status-icon"
                :class="cell.statusClass"
              >
                <component
                  :is="getStatusIcon(cell.status)"
                  class="h-4 w-4"
                />
              </div>
              <div class="status-copy">
                <div class="status-score">
                  {{ cell.statusLabel }}
                </div>
                <div class="status-meta-grid">
                  <div class="status-meta-item">
                    <span class="status-meta-label">Checker</span>
                    <span class="status-meta-value">{{ cell.checkerLabel }}</span>
                  </div>
                  <div class="status-meta-item">
                    <span class="status-meta-label">来源</span>
                    <span class="status-meta-value">{{ cell.sourceLabel }}</span>
                  </div>
                  <div class="status-meta-item">
                    <span class="status-meta-label">状态</span>
                    <span class="status-meta-value">{{ cell.reasonLabel }}</span>
                  </div>
                  <div class="status-meta-item">
                    <span class="status-meta-label">时间</span>
                    <span class="status-meta-value">{{ cell.checkedAtLabel }}</span>
                  </div>
                </div>
              </div>
            </div>
            <div
              v-else
              class="status-empty"
            >
              N/A
            </div>
          </td>
        </tr>
        <tr v-if="rows.length === 0">
          <td :colspan="Math.max(columns.length + 1, 2)">
            <div class="matrix-empty">
              <SearchCheck class="h-5 w-5" />
              <span>{{ emptyMessage }}</span>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
