<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { X } from 'lucide-vue-next'

import type { ContestProjectorFocusPanel } from '@/components/platform/contest/projector/contestProjectorTypes'

const props = defineProps<{
  activePanel: ContestProjectorFocusPanel | null
}>()

const emit = defineEmits<{
  close: []
}>()

function closeOverlay(): void {
  if (!props.activePanel) return
  emit('close')
}

function handleKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Escape') return
  closeOverlay()
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div
    v-if="activePanel"
    class="projector-focus-overlay"
    @click.self="closeOverlay"
  >
    <section
      class="projector-focus-panel"
      role="dialog"
      aria-modal="true"
    >
      <button
        type="button"
        class="projector-focus-close"
        aria-label="关闭聚焦面板"
        title="关闭"
        @click="closeOverlay"
      >
        <X />
      </button>
      <div class="projector-focus-body">
        <slot />
      </div>
    </section>
  </div>
</template>

<style scoped>
.projector-focus-overlay {
  position: fixed;
  z-index: var(--ui-dialog-z-index);
  inset: 0;
  display: grid;
  place-items: center;
  background: var(--ui-dialog-overlay);
  padding: var(--space-6);
}

.projector-focus-panel {
  position: relative;
  width: min(var(--ui-dialog-wide-width), calc(100vw - (var(--space-6) * 2)));
  max-height: calc(100vh - (var(--space-6) * 2));
  overflow: auto;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 30%, var(--color-border-subtle));
  border-radius: var(--ui-dialog-radius);
  background: var(--journal-surface);
  box-shadow: var(--ui-dialog-shadow);
}

.projector-focus-close {
  position: sticky;
  top: var(--space-4);
  float: right;
  display: inline-flex;
  width: var(--ui-control-height-sm);
  height: var(--ui-control-height-sm);
  align-items: center;
  justify-content: center;
  margin: var(--space-4) var(--space-4) 0 0;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--ui-control-radius-sm);
  background: color-mix(in srgb, var(--color-bg-elevated) 88%, transparent);
  color: var(--journal-ink);
  cursor: pointer;
}

.projector-focus-close svg {
  width: var(--space-4);
  height: var(--space-4);
}

.projector-focus-close:hover,
.projector-focus-close:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 36%, var(--color-border-subtle));
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--color-bg-elevated));
}

.projector-focus-close:focus-visible {
  outline: var(--ui-focus-ring-width) solid color-mix(in srgb, var(--journal-accent) 58%, transparent);
  outline-offset: var(--space-1);
}

.projector-focus-body {
  padding: var(--space-5);
}

.projector-focus-body :deep(.leaderboard-panel),
.projector-focus-body :deep(.service-matrix-panel),
.projector-focus-body :deep(.traffic-panel),
.projector-focus-body :deep(.first-blood-panel),
.projector-focus-body :deep(.attack-panel),
.projector-focus-body :deep(.attack-feed-panel) {
  background: color-mix(in srgb, var(--color-bg-elevated) 68%, transparent);
}

@media (max-width: 900px) {
  .projector-focus-overlay {
    padding: var(--space-4);
  }

  .projector-focus-panel {
    width: 100%;
    max-height: calc(100vh - (var(--space-4) * 2));
  }

  .projector-focus-body {
    padding: var(--space-4);
  }
}
</style>
