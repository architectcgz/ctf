<script setup lang="ts">
import type { AWDTrafficSummaryStatView } from './awdTrafficPanel.types'

defineProps<{
  stats: AWDTrafficSummaryStatView[]
}>()
</script>

<template>
  <div class="studio-metric-band">
    <div
      v-for="item in stats"
      :key="item.key"
      class="metric-pill awd-traffic-summary-card"
    >
      <span class="metric-pill__label">{{ item.label }}</span>
      <span class="metric-pill__value">{{ item.value }}</span>
    </div>
  </div>
</template>
