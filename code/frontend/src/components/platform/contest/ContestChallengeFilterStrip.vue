<script setup lang="ts">
import type { ContestChallengePoolFilter } from '@/composables/useContestChallengePool'

defineProps<{
  filterItems: Array<{
    key: ContestChallengePoolFilter
    label: string
    count: number | string
    hint: string
  }>
  activeFilter: ContestChallengePoolFilter
}>()

const emit = defineEmits<{
  select: [key: ContestChallengePoolFilter]
}>()
</script>

<template>
  <div class="contest-challenge-filters">
    <div class="progress-strip metric-panel-grid metric-panel-default-surface">
      <article
        v-for="filter in filterItems"
        :key="filter.key"
        class="journal-note progress-card metric-panel-card"
      >
        <button
          :id="`contest-challenge-filter-${filter.key}`"
          type="button"
          class="contest-challenge-filter"
          :aria-pressed="activeFilter === filter.key"
          @click="emit('select', filter.key)"
        >
          <div class="journal-note-label progress-card-label metric-panel-label">
            {{ filter.label }}
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ filter.count }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            {{ filter.hint }}
          </div>
        </button>
      </article>
    </div>
  </div>
</template>

<style scoped>
.contest-challenge-filters > .progress-strip {
  --metric-panel-columns: repeat(
    auto-fit,
    minmax(min(100%, var(--ui-selector-control-min-width)), 1fr)
  );
}

.contest-challenge-filter {
  display: block;
  width: 100%;
  border: 0;
  background: transparent;
  padding: 0;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}

.contest-challenge-filter:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 72%, transparent);
  outline-offset: var(--space-0-5);
}
</style>
