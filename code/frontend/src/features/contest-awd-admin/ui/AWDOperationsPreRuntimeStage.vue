<script setup lang="ts">
import AWDOperationsTabs from './AWDOperationsTabs.vue'
import type { AWDOperationsPanelKey, AWDOperationsTabItem } from './awdOperations.types'

defineProps<{
  showTabs: boolean
  tabs: readonly AWDOperationsTabItem[]
  activePanel: AWDOperationsPanelKey
  registerTabButton: (key: AWDOperationsPanelKey, element: HTMLButtonElement | null) => void
  contestTitle: string
  hideStudioLink: boolean
  shouldShowRuntimeReadiness: boolean
  runtimeContent: 'all' | 'readiness' | 'round-inspector' | 'instances'
  shouldShowInstanceOrchestration: boolean
}>()

const emit = defineEmits<{
  selectPanel: [panel: AWDOperationsPanelKey]
  tabKeydown: [event: KeyboardEvent, index: number]
  openContestEdit: []
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

    <header class="section-header">
      <div class="section-identity">
        <div class="section-overline">
          Command Center / Pre-flight
        </div>
        <h2 class="section-title">
          {{ contestTitle }}
        </h2>
      </div>
      <div class="section-actions">
        <button
          v-if="!hideStudioLink"
          type="button"
          class="ops-btn ops-btn--neutral"
          @click="emit('openContestEdit')"
        >
          进入竞赛工作室
        </button>
      </div>
    </header>

    <div
      v-if="shouldShowRuntimeReadiness || runtimeContent === 'round-inspector'"
      class="readiness-wrap"
    >
      <slot
        v-if="runtimeContent !== 'readiness'"
        name="pending"
      />
      <slot
        v-if="shouldShowRuntimeReadiness"
        name="readiness"
      />
    </div>

    <slot
      v-if="shouldShowInstanceOrchestration"
      name="instances"
    />
  </section>
</template>

<style scoped>
.studio-ops-section {
  display: flex;
  flex-direction: column;
}

.section-header {
  margin-bottom: var(--space-8);
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-4);
}

.section-overline {
  font-size: var(--font-size-10);
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--color-text-muted);
  margin-bottom: var(--space-1-5);
}

.section-title {
  font-size: var(--font-size-1-25);
  font-weight: 900;
  color: var(--color-text-primary);
  margin: 0;
  letter-spacing: -0.01em;
}

.section-actions {
  display: flex;
  gap: var(--space-3);
}

.ops-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  height: var(--ui-control-height-md);
  padding: 0 var(--space-5);
  border-radius: 0.85rem;
  font-size: var(--font-size-13);
  font-weight: 700;
  transition: all 0.2s ease;
  cursor: pointer;
}

.ops-btn--neutral {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  color: var(--color-text-secondary);
}

.ops-btn--neutral:hover:not(:disabled) {
  border-color: var(--color-primary);
  color: var(--color-text-primary);
}

.ops-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
