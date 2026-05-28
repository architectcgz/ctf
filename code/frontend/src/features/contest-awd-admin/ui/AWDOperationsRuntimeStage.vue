<script setup lang="ts">
import AWDOperationsTabs from './AWDOperationsTabs.vue'
import type { AWDOperationsPanelKey, AWDOperationsTabItem } from './awdOperations.types'

defineProps<{
  showTabs: boolean
  tabs: readonly AWDOperationsTabItem[]
  activePanel: AWDOperationsPanelKey
  registerTabButton: (key: AWDOperationsPanelKey, element: HTMLButtonElement | null) => void
  shouldShowRuntimeReadiness: boolean
  shouldShowRoundInspector: boolean
  shouldShowInstanceOrchestration: boolean
}>()

const emit = defineEmits<{
  selectPanel: [panel: AWDOperationsPanelKey]
  tabKeydown: [event: KeyboardEvent, index: number]
}>()
</script>

<template>
  <section class="studio-ops-section">
    <AWDOperationsTabs
      v-if="showTabs"
      :tabs="tabs"
      :active-panel="activePanel"
      :register-tab-button="registerTabButton"
      @select="emit('selectPanel', $event)"
      @keydown="(event, index) => emit('tabKeydown', event, index)"
    />

    <div class="inspector-wrap">
      <slot
        v-if="shouldShowRuntimeReadiness"
        name="readiness"
      />

      <slot
        v-if="shouldShowRoundInspector"
        name="inspector"
      />

      <slot
        v-if="shouldShowInstanceOrchestration"
        name="instances"
      />
    </div>
  </section>
</template>

<style scoped>
.studio-ops-section {
  display: flex;
  flex-direction: column;
}

.inspector-wrap {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
  min-width: 0;
}

:slotted(.runtime-readiness-strip) {
  margin-bottom: var(--space-6);
}
</style>
