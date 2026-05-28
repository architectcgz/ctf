<script setup lang="ts">
import type { AWDOperationsPanelKey, AWDOperationsTabItem } from './awdOperations.types'

defineProps<{
  tabs: readonly AWDOperationsTabItem[]
  activePanel: AWDOperationsPanelKey
  registerTabButton: (key: AWDOperationsPanelKey, element: HTMLButtonElement | null) => void
}>()

const emit = defineEmits<{
  select: [panel: AWDOperationsPanelKey]
  keydown: [event: KeyboardEvent, index: number]
}>()
</script>

<template>
  <nav class="studio-ops-tabs">
    <button
      v-for="(tab, index) in tabs"
      :id="tab.tabId"
      :key="tab.key"
      :ref="(el) => registerTabButton(tab.key, el as HTMLButtonElement | null)"
      class="tab-item"
      :class="{ active: activePanel === tab.key }"
      type="button"
      role="tab"
      :aria-selected="activePanel === tab.key"
      :aria-controls="tab.panelId"
      @keydown="emit('keydown', $event, index)"
      @click="emit('select', tab.key)"
    >
      {{ tab.label }}
    </button>
  </nav>
</template>

<style scoped>
.studio-ops-tabs {
  display: flex;
  gap: var(--space-8);
  border-bottom: 1px solid var(--color-border-default);
  margin-bottom: var(--space-6);
}

.tab-item {
  padding: var(--space-3) var(--space-1);
  font-size: var(--font-size-13);
  font-weight: 800;
  color: var(--color-text-secondary);
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-item:hover {
  color: var(--color-text-primary);
}

.tab-item.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}
</style>
