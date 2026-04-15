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
    overlayClass?: string | string[]
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
  }>(),
  {
    title: '创建新队伍',
    description: '为你的战队起一个响亮的代号。创建完成后，你可以生成邀请链接让其他队友加入。',
    width: '28.75rem',
    ariaLabel: '专注型输入弹窗',
    overlayClass: '',
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

const resolvedOverlayClass = computed<string[]>(() => {
  if (Array.isArray(props.overlayClass)) {
    return ['c-focused-input-shell', ...props.overlayClass]
  }

  if (props.overlayClass) {
    return ['c-focused-input-shell', props.overlayClass]
  }

  return ['c-focused-input-shell']
})

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
    :overlay-class="resolvedOverlayClass"
    :aria-label="ariaLabel || props.title"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    @update:open="forwardOpen"
    @close="emit('close')"
  >
    <div class="c-focused-input-dialog">
      <div class="c-focused-input-dialog__surface">
        <button
          type="button"
          class="c-focused-input-dialog__close"
          aria-label="关闭弹窗"
          @click="closeDialog"
        >
          <X :size="20" :stroke-width="2" />
        </button>

        <header class="c-focused-input-dialog__header">
          <div class="c-focused-input-dialog__icon">
            <slot name="icon">
              <Users :size="24" :stroke-width="2" />
            </slot>
          </div>

          <h2 class="c-focused-input-dialog__title">{{ props.title }}</h2>
          <p class="c-focused-input-dialog__description">{{ props.description }}</p>
        </header>

        <section class="c-focused-input-dialog__form">
          <slot />
        </section>

        <footer v-if="$slots.footer" class="c-focused-input-dialog__footer">
          <slot name="footer" :close="closeDialog" />
        </footer>
        <footer v-else class="c-focused-input-dialog__footer">
          <button type="button" data-c-modal-action="ghost" @click="closeDialog">取消</button>
          <button type="button" data-c-modal-action="primary">确认创建</button>
        </footer>
      </div>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
.c-focused-input-shell {
  background: rgba(15, 23, 42, 0.3);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.c-focused-input-shell--plain {
  background: transparent;
  backdrop-filter: none;
  -webkit-backdrop-filter: none;
}

.c-focused-input-panel {
  width: min(var(--c-focused-input-width, 28.75rem), calc(100vw - 2rem));
}

.c-focused-input-dialog {
  position: relative;
}

.c-focused-input-dialog__surface {
  position: relative;
  overflow: hidden;
  border-radius: 1rem;
  background: #ffffff;
  box-shadow:
    0 26px 70px rgba(15, 23, 42, 0.24),
    0 8px 24px rgba(15, 23, 42, 0.12);
}

.c-focused-input-dialog__header {
  display: grid;
  justify-items: center;
  gap: 0.9rem;
  padding: 2.25rem 2rem 1.5rem;
  text-align: center;
}

.c-focused-input-dialog__icon {
  display: inline-flex;
  width: 3.5rem;
  height: 3.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  border: 1px solid rgba(16, 185, 129, 0.15);
  background: linear-gradient(180deg, #f2fcf7 0%, #e4f6ed 100%);
  color: #2a7a58;
  box-shadow: 0 10px 24px rgba(42, 122, 88, 0.14);
}

.c-focused-input-dialog__close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  display: inline-flex;
  width: 2.25rem;
  height: 2.25rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgba(248, 250, 252, 0.96);
  color: #94a3b8;
  transition:
    color 0.18s ease,
    background-color 0.18s ease;
}

.c-focused-input-dialog__close:hover {
  color: #475569;
  background: #f1f5f9;
}

.c-focused-input-dialog__title {
  margin: 0;
  font-size: 1.9rem;
  font-weight: 700;
  line-height: 1.1;
  color: #0f172a;
}

.c-focused-input-dialog__description {
  margin: 0;
  max-width: 27ch;
  font-size: 14px;
  line-height: 1.7;
  color: #475569;
}

.c-focused-input-dialog__form {
  display: grid;
  gap: 1.25rem;
  padding: 0 2rem 1.75rem;
}

.c-focused-input-dialog__form :deep(label) {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 12px;
  font-weight: 700;
  color: #334155;
}

.c-focused-input-dialog__form :deep(input),
.c-focused-input-dialog__form :deep(textarea),
.c-focused-input-dialog__form :deep(select) {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.8rem;
  background: #f8fafc;
  font-size: 14px;
  outline: none;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.45);
}

.c-focused-input-dialog__form :deep(input::placeholder),
.c-focused-input-dialog__form :deep(textarea::placeholder) {
  color: #94a3b8;
}

.c-focused-input-dialog__form :deep(input:focus),
.c-focused-input-dialog__form :deep(textarea:focus),
.c-focused-input-dialog__form :deep(select:focus) {
  border-color: #2a7a58;
  background: #ffffff;
  box-shadow: 0 0 0 4px rgba(42, 122, 88, 0.1);
}

.c-focused-input-dialog__footer {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 0 2rem 2rem;
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='ghost']),
.c-focused-input-dialog__footer :deep([data-c-modal-action='primary']) {
  min-height: 2.75rem;
  border-radius: 0.8rem;
  font-size: 14px;
  font-weight: 600;
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='ghost']) {
  padding: 0.7rem 1.2rem;
  color: #475569;
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='ghost']:hover) {
  background: #f1f5f9;
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='primary']) {
  padding: 0.7rem 1.5rem;
  background: #2a7a58;
  color: #ffffff;
  box-shadow: 0 12px 24px rgba(42, 122, 88, 0.2);
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='primary']:hover) {
  background: #206346;
  transform: translateY(-1px);
}

@media (max-width: 640px) {
  .c-focused-input-dialog__header {
    padding: 2rem 1.25rem 1.35rem;
  }

  .c-focused-input-dialog__form,
  .c-focused-input-dialog__footer {
    padding-inline: 1.25rem;
  }

  .c-focused-input-dialog__footer {
    justify-content: stretch;
    padding-bottom: 1.25rem;
  }

  .c-focused-input-dialog__footer :deep([data-c-modal-action='ghost']),
  .c-focused-input-dialog__footer :deep([data-c-modal-action='primary']) {
    flex: 1 1 100%;
    justify-content: center;
  }
}
</style>
