<script setup lang="ts">
import { computed } from 'vue'
import { AlignLeft, X } from 'lucide-vue-next'

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
    title: '题目高级配置',
    subtitle: '在这里放置长表单、复杂配置或带有高度承载需求的内容。',
    eyebrow: 'Advanced Editor',
    width: '32rem',
    closeOnBackdrop: true,
    closeOnEscape: true,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()

const overlayClass = 'modal-template-shell--drawer'
// 将布局参数和视觉参数统一通过 CSS 变量传递
const panelStyle = computed<Record<string, string>>(() => ({
  '--modal-template-drawer-width': props.width,
  '--drawer-overlay-blur': '12px',
  '--drawer-overlay-bg': 'color-mix(in srgb, var(--color-bg-base) 45%, transparent)',
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
    panel-class="modal-template-panel--drawer"
    panel-tag="aside"
    :aria-label="title"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    :style="{
      '--modal-shell-justify': 'flex-end',
      '--modal-shell-align': 'stretch',
      '--modal-shell-padding': '0',
      '--modal-shell-blur': '12px'
    }"
    @update:open="forwardOpen"
    @close="forwardClose"
  >
    <div class="modal-template-drawer">
      <header class="modal-template-drawer__header">
        <div class="modal-template-drawer__head-row">
          <div class="modal-template-drawer__icon">
            <slot name="icon">
              <AlignLeft class="h-5 w-5" />
            </slot>
          </div>
          <button
            type="button"
            class="modal-template-drawer__close"
            aria-label="关闭抽屉"
            @click="forwardClose(), forwardOpen(false)"
          >
            <X class="h-4 w-4" />
          </button>
        </div>

        <p
          v-if="eyebrow"
          class="modal-template-drawer__eyebrow"
        >
          {{ eyebrow }}
        </p>
        <h2 class="modal-template-drawer__title">
          {{ title }}
        </h2>
        <p
          v-if="subtitle"
          class="modal-template-drawer__subtitle"
        >
          {{ subtitle }}
        </p>
      </header>

      <div class="modal-template-drawer__body">
        <slot />
      </div>

      <footer
        v-if="$slots.footer"
        class="modal-template-drawer__footer"
      >
        <slot name="footer" />
      </footer>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
/* 
  优雅的做法：
  1. 通过变量继承影响 Shell 组件。
  2. 局部定义抽屉内部的语义色，不再污染全局。
*/
.modal-template-shell--drawer {
  --modal-template-drawer-accent: var(--color-primary);
  --modal-template-drawer-line: var(--color-border-subtle);
}

:deep(.modal-template-panel--drawer) {
  width: var(--modal-template-drawer-width);
  max-width: 100%;
}

.modal-template-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
  /* 确保不透明背景 */
  background-color: var(--color-bg-surface);
}

.modal-template-drawer__header {
  padding: 1.75rem 2rem 1.5rem;
  border-bottom: 1px solid var(--modal-template-drawer-line);
}

.modal-template-drawer__head-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.modal-template-drawer__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--modal-template-drawer-accent) 18%, var(--modal-template-drawer-line));
  background: color-mix(in srgb, var(--modal-template-drawer-accent) 10%, var(--color-bg-surface));
  color: color-mix(in srgb, var(--modal-template-drawer-accent) 92%, var(--color-text-primary));
  box-shadow: 0 8px 20px color-mix(in srgb, var(--modal-template-drawer-accent) 10%, transparent);
}

.modal-template-drawer__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid var(--modal-template-drawer-line);
  border-radius: 999px;
  background: var(--color-bg-elevated);
  color: var(--color-text-muted);
  transition: all 0.18s ease;
}

.modal-template-drawer__close:hover {
  background: var(--color-border-default);
  color: var(--color-text-primary);
}

.modal-template-drawer__eyebrow {
  margin: 1rem 0 0;
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.modal-template-drawer__title {
  margin: 0.55rem 0 0;
  font-size: 1.6rem;
  font-weight: 900;
  line-height: 1.08;
  color: var(--color-text-primary);
}

.modal-template-drawer__subtitle {
  margin: 0.75rem 0 0;
  font-size: 0.8rem;
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.modal-template-drawer__body {
  flex: 1 1 auto;
  overflow-y: auto;
  padding: 2rem;
}

.modal-template-drawer__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid var(--modal-template-drawer-line);
  background-color: var(--color-bg-surface);
  flex-shrink: 0;
}
</style>
