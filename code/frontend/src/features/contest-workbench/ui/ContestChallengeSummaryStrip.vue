<script setup lang="ts">
defineProps<{
  summaryItems: Array<{
    key: string
    label: string
    value: number | string
    hint: string
  }>
}>()
</script>

<template>
  <div class="progress-strip metric-panel-grid metric-panel-default-surface contest-challenge-panel__summary">
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
</template>

<style scoped>
.contest-challenge-panel__summary {
  --admin-summary-grid-columns: repeat(auto-fit, minmax(11rem, 1fr));
  --metric-panel-border: color-mix(in srgb, var(--workspace-line-soft) 86%, transparent);
  --metric-panel-shadow: 0 var(--space-2) var(--space-5)
    color-mix(in srgb, var(--color-shadow-soft) 24%, transparent);
}

@media (max-width: 900px) {
  .contest-challenge-panel__summary {
    --admin-summary-grid-columns: minmax(0, 1fr);
  }
}
</style>
