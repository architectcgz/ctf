<script setup lang="ts">
import type { ContestWorkbenchSummaryItem } from '@/composables/useContestWorkbench'

defineProps<{
  items: ContestWorkbenchSummaryItem[]
}>()
</script>

<template>
  <section
    class="metric-panel-grid metric-panel-default-surface contest-workbench-summary-strip"
    aria-label="竞赛工作台摘要"
  >
    <article
      v-for="item in items"
      :key="item.key"
      class="journal-note metric-panel-card contest-workbench-summary-strip__item"
    >
      <div class="journal-note-label metric-panel-label">{{ item.label }}</div>
      <div class="journal-note-value metric-panel-value">{{ item.value }}</div>
      <div v-if="item.hint" class="journal-note-helper metric-panel-helper">{{ item.hint }}</div>
    </article>
  </section>
</template>

<style scoped>
.contest-workbench-summary-strip {
  gap: var(--space-3);
}

.contest-workbench-summary-strip__item {
  min-width: 0;
}
</style>
