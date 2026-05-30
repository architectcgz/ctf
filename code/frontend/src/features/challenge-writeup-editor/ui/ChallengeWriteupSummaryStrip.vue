<script setup lang="ts">
import { FileText, Users } from 'lucide-vue-next'

defineProps<{
  officialWriteupCount: number
  submissionTotal: number
}>()
</script>

<template>
  <div class="writeup-manage-stats-shell">
    <div
      class="admin-summary-grid writeup-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
    >
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>官方题解</span>
          <FileText class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ officialWriteupCount }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          当前题目已创建的官方题解数量
        </div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>学员题解</span>
          <Users class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ submissionTotal }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          当前题目收到的学员题解投稿数量
        </div>
      </article>
    </div>
  </div>
</template>
