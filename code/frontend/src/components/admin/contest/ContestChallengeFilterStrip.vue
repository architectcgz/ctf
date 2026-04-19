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
    <button
      v-for="filter in filterItems"
      :id="`contest-challenge-filter-${filter.key}`"
      :key="filter.key"
      type="button"
      class="contest-challenge-filter"
      :class="{ 'contest-challenge-filter--active': activeFilter === filter.key }"
      @click="emit('select', filter.key)"
    >
      <span class="contest-challenge-filter__label">{{ filter.label }}</span>
      <span class="contest-challenge-filter__count">{{ filter.count }}</span>
      <span class="contest-challenge-filter__hint">{{ filter.hint }}</span>
    </button>
  </div>
</template>

<style scoped>
.contest-challenge-filters {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(auto-fit, minmax(10.5rem, 1fr));
}

.contest-challenge-filter {
  display: grid;
  gap: var(--space-1);
  justify-items: start;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: var(--space-3);
  text-align: left;
  transition: all 150ms ease;
}

.contest-challenge-filter:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
}

.contest-challenge-filter--active {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
}

.contest-challenge-filter__label,
.contest-challenge-filter__count {
  font-weight: 700;
  color: var(--journal-ink);
}

.contest-challenge-filter__count {
  font-size: var(--font-size-1-20);
}

.contest-challenge-filter__hint {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}
</style>
