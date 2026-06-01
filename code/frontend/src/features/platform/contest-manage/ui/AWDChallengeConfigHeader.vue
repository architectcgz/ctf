<script setup lang="ts">
import type { AWDChallengeConfigSummaryItem } from './awdChallengeConfigPanel.types'

defineProps<{
  summaryItems: AWDChallengeConfigSummaryItem[]
}>()
</script>

<template>
  <header class="studio-pane-header">
    <div class="header-main">
      <div class="workspace-overline">
        AWD Service Config
      </div>
      <h1 class="pane-title">
        AWD 编排
      </h1>
      <p class="pane-description">
        维护 Checker、SLA / 防守权重、试跑结果和就绪状态。
      </p>
    </div>

    <div class="progress-strip metric-panel-grid metric-panel-default-surface">
      <article
        v-for="item in summaryItems"
        :key="item.key"
        class="journal-note progress-card metric-panel-card"
      >
        <div class="journal-note-label progress-card-label metric-panel-label">
          {{ item.label }}
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ item.value }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          {{ item.hint }}
        </div>
      </article>
    </div>
  </header>
</template>
