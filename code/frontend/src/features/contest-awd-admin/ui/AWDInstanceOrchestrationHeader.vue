<script setup lang="ts">
import { Play, RefreshCw, ShieldCheck } from 'lucide-vue-next'

import type { AWDInstanceOrchestrationHeaderView } from './awdInstanceOrchestration.types'

defineProps<{
  summary: AWDInstanceOrchestrationHeaderView
}>()

const emit = defineEmits<{
  refresh: []
  startAll: []
}>()
</script>

<template>
  <header class="orchestration-header">
    <div class="orchestration-heading">
      <h3 class="orchestration-title">
        实例编排
      </h3>
    </div>
    <div class="orchestration-actions">
      <div class="orchestration-summary">
        <ShieldCheck class="summary-icon" />
        <span>{{ summary.runningCount }} / {{ summary.totalTargetCount }}</span>
      </div>
      <button
        type="button"
        class="ops-btn ops-btn--neutral"
        :disabled="summary.loading || summary.hasPendingStart"
        title="刷新实例编排"
        @click="emit('refresh')"
      >
        <RefreshCw
          class="btn-icon"
          :class="{ 'animate-spin': summary.loading }"
        />
      </button>
      <button
        type="button"
        class="ops-btn ops-btn--primary"
        :disabled="!summary.canStartAll"
        @click="emit('startAll')"
      >
        <Play class="btn-icon" />
        <span>启动全部</span>
      </button>
    </div>
  </header>
</template>
