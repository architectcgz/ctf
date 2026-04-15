<script setup lang="ts">
import { computed } from 'vue'
import { Edit3, X } from 'lucide-vue-next'

import ModalTemplateShell from './ModalTemplateShell.vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    title: string
    subtitle?: string
    eyebrow?: string
    width?: string
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
  }>(),
  {
    subtitle: '',
    eyebrow: 'Resource Editor',
    width: '32rem',
    closeOnBackdrop: true,
    closeOnEscape: true,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()

const panelStyle = computed<Record<string, string>>(() => ({
  '--modal-template-classic-width': props.width,
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
    panel-class="modal-template-panel--classic"
    :aria-label="title"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    @update:open="forwardOpen"
    @close="forwardClose"
  >
    <div class="modal-template-classic">
      <header class="modal-template-classic__header">
        <div class="modal-template-classic__identity">
          <div class="modal-template-classic__icon">
            <slot name="icon">
              <Edit3 class="h-4 w-4" />
            </slot>
          </div>
          <div class="modal-template-classic__copy">
            <p v-if="eyebrow" class="modal-template-classic__eyebrow">{{ eyebrow }}</p>
            <h2 class="modal-template-classic__title">{{ title }}</h2>
            <p v-if="subtitle" class="modal-template-classic__subtitle">{{ subtitle }}</p>
          </div>
        </div>

        <button
          type="button"
          class="modal-template-classic__close"
          aria-label="关闭弹窗"
          @click="forwardClose(), forwardOpen(false)"
        >
          <X class="h-4.5 w-4.5" />
        </button>
      </header>

      <div class="modal-template-classic__body">
        <slot />
      </div>

      <footer v-if="$slots.footer" class="modal-template-classic__footer">
        <slot name="footer" />
      </footer>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
.modal-template-panel--classic {
  width: min(var(--modal-template-classic-width, 32rem), 100%);
  border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 1.5rem;
  background: #ffffff;
  box-shadow: 0 28px 80px rgba(15, 23, 42, 0.22);
  overflow: hidden;
}

.modal-template-classic__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.35rem 1.5rem 1.2rem;
  border-bottom: 1px solid rgba(241, 245, 249, 1);
  background: rgba(248, 250, 252, 0.68);
}

.modal-template-classic__identity {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  min-width: 0;
}

.modal-template-classic__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: 0.85rem;
  background: rgba(219, 234, 254, 1);
  color: #2563eb;
  flex-shrink: 0;
}

.modal-template-classic__copy {
  min-width: 0;
}

.modal-template-classic__eyebrow {
  margin: 0;
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: rgba(148, 163, 184, 1);
}

.modal-template-classic__title {
  margin: 0.25rem 0 0;
  font-size: 0.95rem;
  font-weight: 900;
  line-height: 1.2;
  color: rgba(15, 23, 42, 1);
}

.modal-template-classic__subtitle {
  margin: 0.35rem 0 0;
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: rgba(148, 163, 184, 1);
}

.modal-template-classic__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border: 0;
  border-radius: 0.75rem;
  background: transparent;
  color: rgba(148, 163, 184, 1);
  transition:
    background-color 0.18s ease,
    color 0.18s ease;
}

.modal-template-classic__close:hover {
  background: rgba(241, 245, 249, 1);
  color: rgba(51, 65, 85, 1);
}

.modal-template-classic__body {
  padding: 1.5rem;
}

.modal-template-classic__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem 1.25rem;
  border-top: 1px solid rgba(241, 245, 249, 1);
  background: rgba(248, 250, 252, 0.68);
}
</style>
