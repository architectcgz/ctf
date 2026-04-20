<script setup lang="ts">
import type { AwdHeroMetric } from '@/modules/awd/types'

defineProps<{
  metrics: AwdHeroMetric[]
}>()
</script>

<template>
  <div
    v-if="metrics.length > 0"
    class="awd-metric-strip metric-panel-grid"
  >
    <article
      v-for="metric in metrics"
      :key="metric.label"
      class="metric-panel-card awd-metric-strip__item"
    >
      <div class="metric-panel-label">
        {{ metric.label }}
      </div>
      <div class="metric-panel-value">
        {{ metric.value }}
      </div>
      <div
        v-if="metric.helper"
        class="metric-panel-helper"
      >
        {{ metric.helper }}
      </div>
    </article>
  </div>
</template>

<style scoped>
.awd-metric-strip {
  --metric-panel-columns: repeat(auto-fit, minmax(170px, 1fr));
}

.awd-metric-strip__item {
  min-height: 100%;
}
</style>
