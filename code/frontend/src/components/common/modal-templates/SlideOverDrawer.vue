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
const panelStyle = computed<Record<string, string>>(() => ({
  '--modal-template-drawer-width': props.width,
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

        <p v-if="eyebrow" class="modal-template-drawer__eyebrow">{{ eyebrow }}</p>
        <h2 class="modal-template-drawer__title">{{ title }}</h2>
        <p v-if="subtitle" class="modal-template-drawer__subtitle">{{ subtitle }}</p>
      </header>

      <div class="modal-template-drawer__body">
        <slot />
      </div>

      <footer v-if="$slots.footer" class="modal-template-drawer__footer">
        <slot name="footer" />
      </footer>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
.modal-template-shell--drawer {
  justify-content: flex-end;
  padding: 0;
  background: rgba(15, 23, 42, 0.2);
  backdrop-filter: blur(8px);
}

.modal-template-panel--drawer {
  width: min(var(--modal-template-drawer-width, 32rem), 100vw);
  height: 100%;
  border-left: 1px solid rgba(226, 232, 240, 1);
  background: #ffffff;
  box-shadow: -24px 0 64px rgba(15, 23, 42, 0.16);
}

.modal-template-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.modal-template-drawer__header {
  padding: 1.75rem 2rem 1.5rem;
  border-bottom: 1px solid rgba(241, 245, 249, 1);
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
  border: 1px solid rgba(209, 250, 229, 1);
  background: rgba(236, 253, 245, 1);
  color: #059669;
  box-shadow: 0 8px 20px rgba(16, 185, 129, 0.08);
}

.modal-template-drawer__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid rgba(226, 232, 240, 1);
  border-radius: 999px;
  background: rgba(248, 250, 252, 1);
  color: rgba(148, 163, 184, 1);
  transition:
    background-color 0.18s ease,
    color 0.18s ease;
}

.modal-template-drawer__close:hover {
  background: rgba(241, 245, 249, 1);
  color: rgba(51, 65, 85, 1);
}

.modal-template-drawer__eyebrow {
  margin: 1rem 0 0;
  font-size: 0.625rem;
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: rgba(148, 163, 184, 1);
}

.modal-template-drawer__title {
  margin: 0.55rem 0 0;
  font-size: 1.6rem;
  font-weight: 900;
  line-height: 1.08;
  color: rgba(15, 23, 42, 1);
}

.modal-template-drawer__subtitle {
  margin: 0.75rem 0 0;
  font-size: 0.8rem;
  line-height: 1.7;
  color: rgba(100, 116, 139, 1);
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
  border-top: 1px solid rgba(226, 232, 240, 1);
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(14px);
  flex-shrink: 0;
}
</style>
