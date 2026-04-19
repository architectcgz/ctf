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
    frosted?: boolean
  }>(),
  {
    subtitle: '',
    eyebrow: 'Resource Editor',
    width: '32rem',
    closeOnBackdrop: true,
    closeOnEscape: true,
    frosted: false,
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
    :frosted="frosted"
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
/* 
 * 这里只保留【结构性】样式。
 * 视觉风格（背景、边框、毛玻璃）已移至全局 style.css 以解决 Teleport 样式隔离问题。
 */
.modal-template-panel--classic {
  width: min(var(--modal-template-classic-width, 32rem), 100%);
  display: flex;
  flex-direction: column;
}

.modal-template-classic__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.5rem 1.75rem 1.25rem;
}

.modal-template-classic__identity {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  min-width: 0;
}

.modal-template-classic__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.95rem;
  /* 图标背景仍保留少许局部逻辑，或者可移出 */
  background: color-mix(in srgb, var(--color-primary) 10%, var(--color-bg-surface));
  color: color-mix(in srgb, var(--color-primary) 92%, var(--color-text-primary));
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
  color: var(--color-text-muted);
}

.modal-template-classic__title {
  margin: 0.25rem 0 0;
  font-size: 1rem;
  font-weight: 900;
  line-height: 1.2;
  color: var(--color-text-primary);
}

.modal-template-classic__subtitle {
  margin: 0.35rem 0 0;
  font-size: 0.75rem;
  font-weight: 500;
  line-height: 1.5;
  color: var(--color-text-secondary);
}

.modal-template-classic__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 0;
  border-radius: 0.85rem;
  background: transparent;
  color: var(--color-text-muted);
  transition: all 0.18s ease;
}

.modal-template-classic__close:hover {
  background: color-mix(in srgb, var(--color-text-primary) 8%, transparent);
  color: var(--color-text-primary);
}

.modal-template-classic__body {
  padding: 1.75rem;
}

.modal-template-classic__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.25rem 1.75rem 1.5rem;
}
</style>
