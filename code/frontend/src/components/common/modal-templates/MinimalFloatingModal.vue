<script setup lang="ts">
import { computed } from 'vue'
import { X } from 'lucide-vue-next'

import ModalTemplateShell from './ModalTemplateShell.vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    width?: string
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
  }>(),
  {
    title: '快捷编辑',
    width: '25rem',
    closeOnBackdrop: true,
    closeOnEscape: true,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()

const overlayClass = 'modal-template-shell--minimal'
const panelStyle = computed<Record<string, string>>(() => ({
  '--modal-template-minimal-width': props.width,
}))

function forwardOpen(value: boolean): void {
  emit('update:open', value)
}

function forwardClose(): void {
  emit('close')
}
</script>

<template>
  <ModalTemplateShell
    :open="open"
    :panel-style="panelStyle"
    :overlay-class="overlayClass"
    panel-class="modal-template-panel--minimal"
    :aria-label="title"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    @update:open="forwardOpen"
    @close="forwardClose"
  >
    <div class="modal-template-minimal">
      <header class="modal-template-minimal__header">
        <h2 class="modal-template-minimal__title">{{ title }}</h2>
        <button
          type="button"
          class="modal-template-minimal__close"
          aria-label="关闭浮窗"
          @click="forwardClose(), forwardOpen(false)"
        >
          <X class="h-4.5 w-4.5" />
        </button>
      </header>

      <div class="modal-template-minimal__body">
        <slot />
      </div>

      <footer v-if="$slots.footer" class="modal-template-minimal__footer">
        <slot name="footer" />
      </footer>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
.modal-template-shell--minimal {
  --modal-template-minimal-surface: color-mix(in srgb, var(--color-bg-elevated) 96%, var(--color-bg-surface));
  --modal-template-minimal-line: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --modal-template-minimal-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --modal-template-minimal-muted: color-mix(in srgb, var(--color-text-secondary) 92%, transparent);
  --modal-template-minimal-faint: color-mix(in srgb, var(--color-text-muted) 92%, transparent);
  --modal-template-minimal-accent: var(--color-primary);
  background: transparent;
}

.modal-template-panel--minimal {
  width: min(var(--modal-template-minimal-width, 25rem), 100%);
  border: 1px solid var(--modal-template-minimal-line);
  border-radius: 1.5rem;
  background: var(--modal-template-minimal-surface);
  box-shadow: 0 12px 40px color-mix(in srgb, var(--color-shadow-strong) 16%, transparent);
}

.modal-template-minimal {
  padding: 1.5rem;
}

.modal-template-minimal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.modal-template-minimal__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.01em;
  color: var(--modal-template-minimal-text);
}

.modal-template-minimal__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: 0;
  background: transparent;
  color: var(--modal-template-minimal-faint);
  transition: color 0.18s ease;
}

.modal-template-minimal__close:hover {
  color: var(--modal-template-minimal-muted);
}

.modal-template-minimal__body {
  color: var(--modal-template-minimal-text);
}

.modal-template-minimal__footer {
  margin-top: 2rem;
}

.modal-template-minimal :deep(input),
.modal-template-minimal :deep(select),
.modal-template-minimal :deep(textarea) {
  border-bottom-color: var(--modal-template-minimal-line);
}

.modal-template-minimal :deep(input:focus),
.modal-template-minimal :deep(select:focus),
.modal-template-minimal :deep(textarea:focus) {
  border-bottom-color: color-mix(in srgb, var(--modal-template-minimal-accent) 92%, var(--modal-template-minimal-text));
}
</style>
