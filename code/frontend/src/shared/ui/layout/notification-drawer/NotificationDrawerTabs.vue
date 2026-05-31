<template>
  <section class="tabs" aria-label="通知筛选">
    <button
      v-for="filter in filterOptions"
      :key="filter.value"
      type="button"
      class="tab-btn"
      :class="{ 'is-active': activeFilter === filter.value }"
      :aria-pressed="activeFilter === filter.value"
      @click="$emit('select', filter.value)"
    >
      {{ filter.label }}
    </button>
  </section>
</template>

<script setup lang="ts">
import type { NotificationFilter, NotificationFilterOption } from '@/shared/model/layout/notification-drawer/types'

defineProps<{
  activeFilter: NotificationFilter
  filterOptions: NotificationFilterOption[]
}>()

defineEmits<{
  select: [value: NotificationFilter]
}>()
</script>

<style scoped>
.tabs {
  margin-top: var(--space-6);
  min-height: 2.25rem;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  align-items: center;
  column-gap: var(--space-1-5);
  padding: var(--space-0-5);
  border: 1px solid var(--notification-tab-shell-border);
  border-radius: var(--ui-control-radius-lg);
  background: var(--notification-tab-shell-bg);
}

.tab-btn {
  height: 1.75rem;
  border: 1px solid var(--notification-tab-border);
  border-radius: var(--ui-control-radius-md);
  background: var(--notification-tab-bg);
  color: var(--notification-tab-text);
  font-size: var(--font-size-13);
  font-weight: 500;
  cursor: pointer;
  transition:
    border-color var(--ui-motion-fast),
    background-color var(--ui-motion-fast),
    color var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast);
}

.tab-btn:hover:not(.is-active) {
  color: var(--notification-tab-hover-text);
  background: var(--notification-tab-hover-bg);
}

.tab-btn.is-active {
  color: var(--notification-tab-active-text);
  font-weight: 600;
  background: var(--notification-tab-active-bg);
  border-color: var(--notification-tab-active-border);
  box-shadow: var(--notification-tab-active-shadow);
}

.tab-btn:focus-visible {
  outline: var(--ui-focus-ring-width) solid rgb(114 184 255 / 0.46);
  outline-color: var(--color-primary);
  outline-offset: calc(var(--space-0-5) * -1);
}

@media (max-width: 768px) {
  .tabs {
    margin-top: var(--space-5);
  }
}
</style>
