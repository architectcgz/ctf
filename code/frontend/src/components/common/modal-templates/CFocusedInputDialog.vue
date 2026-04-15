<script setup lang="ts">
import { computed } from 'vue'
import { Users, X } from 'lucide-vue-next'

import ModalTemplateShell from './ModalTemplateShell.vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    description?: string
    width?: string
    ariaLabel?: string
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
  }>(),
  {
    title: '创建新队伍',
    description: '为你的战队起一个响亮的代号。创建完成后，你可以生成邀请链接让其他队友加入。',
    width: '27.5rem',
    ariaLabel: '专注型输入弹窗',
    closeOnBackdrop: true,
    closeOnEscape: true,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()

const panelStyle = computed<Record<string, string>>(() => ({
  '--c-focused-input-width': props.width,
}))

function forwardOpen(value: boolean): void {
  emit('update:open', value)
}

function closeDialog(): void {
  emit('close')
  emit('update:open', false)
}
</script>

<template>
  <ModalTemplateShell
    :open="open"
    :panel-style="panelStyle"
    panel-class="c-focused-input-panel"
    overlay-class="c-focused-input-shell"
    :aria-label="ariaLabel || props.title"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    @update:open="forwardOpen"
    @close="emit('close')"
  >
    <div class="c-focused-input-dialog">
      <div class="c-focused-input-dialog__head">
        <div class="c-focused-input-dialog__icon">
          <slot name="icon">
            <Users :size="24" :stroke-width="2" />
          </slot>
        </div>
        <button
          type="button"
          class="c-focused-input-dialog__close"
          aria-label="关闭弹窗"
          @click="closeDialog"
        >
          <X :size="20" :stroke-width="2" />
        </button>
      </div>

      <h2 class="c-focused-input-dialog__title">{{ props.title }}</h2>
      <p class="c-focused-input-dialog__description">{{ props.description }}</p>

      <div class="c-focused-input-dialog__body">
        <slot />
      </div>

      <footer v-if="$slots.footer" class="c-focused-input-dialog__footer">
        <slot name="footer" :close="closeDialog" />
      </footer>
      <footer v-else class="c-focused-input-dialog__footer">
        <button type="button" data-c-modal-action="ghost" @click="closeDialog">取消</button>
        <button type="button" data-c-modal-action="primary">确认创建</button>
      </footer>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
.c-focused-input-shell {
  background: rgba(15, 23, 42, 0.3);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.c-focused-input-panel {
  width: min(var(--c-focused-input-width, 27.5rem), 100%);
  overflow: hidden;
  border-radius: 0.25rem;
  background: #ffffff;
  box-shadow: 0 25px 50px rgba(15, 23, 42, 0.28);
}

.c-focused-input-dialog {
  padding: 2rem;
}

.c-focused-input-dialog__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.c-focused-input-dialog__icon {
  display: inline-flex;
  width: 3rem;
  height: 3rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  border: 1px solid rgba(16, 185, 129, 0.15);
  background: #f2fcf7;
  color: #2a7a58;
}

.c-focused-input-dialog__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  transition: color 0.18s ease;
}

.c-focused-input-dialog__close:hover {
  color: #475569;
}

.c-focused-input-dialog__title {
  margin: 0;
  font-size: 2rem;
  font-weight: 700;
  color: #0f172a;
}

.c-focused-input-dialog__description {
  margin: 0.5rem 0 2rem;
  font-size: 14px;
  line-height: 1.7;
  color: #475569;
}

.c-focused-input-dialog__body :deep(label) {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 12px;
  font-weight: 700;
  color: #334155;
}

.c-focused-input-dialog__body :deep(input),
.c-focused-input-dialog__body :deep(textarea),
.c-focused-input-dialog__body :deep(select) {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  background: #f8fafc;
  font-size: 14px;
  outline: none;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.c-focused-input-dialog__body :deep(input::placeholder),
.c-focused-input-dialog__body :deep(textarea::placeholder) {
  color: #94a3b8;
}

.c-focused-input-dialog__body :deep(input:focus),
.c-focused-input-dialog__body :deep(textarea:focus),
.c-focused-input-dialog__body :deep(select:focus) {
  border-color: #2a7a58;
  background: #ffffff;
  box-shadow: 0 0 0 4px rgba(42, 122, 88, 0.1);
}

.c-focused-input-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 0.5rem;
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='ghost']),
.c-focused-input-dialog__footer :deep([data-c-modal-action='primary']) {
  border-radius: 0.25rem;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='ghost']) {
  padding: 0.625rem 1.25rem;
  color: #475569;
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='ghost']:hover) {
  background: #f1f5f9;
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='primary']) {
  padding: 0.625rem 1.5rem;
  background: #2a7a58;
  color: #ffffff;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.12);
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='primary']:hover) {
  background: #206346;
}
</style>
