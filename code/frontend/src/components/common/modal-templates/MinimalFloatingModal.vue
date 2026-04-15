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
  background: transparent;
}

.modal-template-panel--minimal {
  width: min(var(--modal-template-minimal-width, 25rem), 100%);
  border: 1px solid rgba(226, 232, 240, 0.85);
  border-radius: 1.5rem;
  background: #ffffff;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.12);
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
  color: rgba(30, 41, 59, 1);
}

.modal-template-minimal__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: 0;
  background: transparent;
  color: rgba(148, 163, 184, 1);
  transition: color 0.18s ease;
}

.modal-template-minimal__close:hover {
  color: rgba(51, 65, 85, 1);
}

.modal-template-minimal__body {
  color: rgba(15, 23, 42, 1);
}

.modal-template-minimal__footer {
  margin-top: 2rem;
}

.modal-template-minimal :deep(input),
.modal-template-minimal :deep(select),
.modal-template-minimal :deep(textarea) {
  border-bottom-color: rgba(226, 232, 240, 1);
}

.modal-template-minimal :deep(input:focus),
.modal-template-minimal :deep(select:focus),
.modal-template-minimal :deep(textarea:focus) {
  border-bottom-color: #7c3aed;
}
</style>
