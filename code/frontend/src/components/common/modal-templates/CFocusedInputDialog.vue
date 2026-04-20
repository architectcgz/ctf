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
          <X
            :size="20"
            :stroke-width="2"
          />
        </button>

        <header class="c-focused-input-dialog__header">
          <div class="c-focused-input-dialog__icon">
            <slot name="icon">
              <Users
                :size="24"
                :stroke-width="2"
              />
            </slot>
          </div>

          <h2 class="c-focused-input-dialog__title">
            {{ props.title }}
          </h2>
          <p class="c-focused-input-dialog__description">
            {{ props.description }}
          </p>
        </header>

        <section class="c-focused-input-dialog__form">
          <slot />
        </section>

        <footer
          v-if="$slots.footer"
          class="c-focused-input-dialog__footer"
        >
          <slot
            name="footer"
            :close="closeDialog"
          />
        </footer>
        <footer
          v-else
          class="c-focused-input-dialog__footer"
        >
          <button
            type="button"
            data-c-modal-action="ghost"
            @click="closeDialog"
          >
            取消
          </button>
          <button
            type="button"
            data-c-modal-action="primary"
          >
            确认创建
          </button>
        </footer>
      </div>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
.c-focused-input-shell {
  --c-focused-input-overlay: color-mix(in srgb, var(--color-bg-base) 28%, transparent);
  --c-focused-input-surface: color-mix(in srgb, var(--color-bg-elevated) 96%, var(--color-bg-surface));
  --c-focused-input-surface-muted: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --c-focused-input-line: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --c-focused-input-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --c-focused-input-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --c-focused-input-muted: color-mix(in srgb, var(--color-text-secondary) 92%, transparent);
  --c-focused-input-faint: color-mix(in srgb, var(--color-text-muted) 94%, transparent);
  --c-focused-input-accent: var(--color-primary);
  background: var(--c-focused-input-overlay);
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
  background: var(--c-focused-input-surface);
  box-shadow:
    0 26px 70px color-mix(in srgb, var(--color-shadow-strong) 26%, transparent),
    0 8px 24px color-mix(in srgb, var(--color-shadow-soft) 24%, transparent);
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
  border: 1px solid color-mix(in srgb, var(--c-focused-input-accent) 18%, var(--c-focused-input-line));
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--c-focused-input-accent) 14%, var(--c-focused-input-surface)),
    color-mix(in srgb, var(--c-focused-input-accent) 6%, var(--c-focused-input-surface-muted))
  );
  color: color-mix(in srgb, var(--c-focused-input-accent) 92%, var(--c-focused-input-text));
  box-shadow: 0 10px 24px color-mix(in srgb, var(--c-focused-input-accent) 14%, transparent);
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
  background: color-mix(in srgb, var(--c-focused-input-surface-muted) 96%, transparent);
  color: var(--c-focused-input-faint);
  transition:
    color 0.18s ease,
    background-color 0.18s ease;
}

.c-focused-input-dialog__close:hover {
  color: var(--c-focused-input-muted);
  background: color-mix(in srgb, var(--c-focused-input-line) 18%, var(--c-focused-input-surface-muted));
}

.c-focused-input-dialog__title {
  margin: 0;
  font-size: 1.9rem;
  font-weight: 700;
  line-height: 1.1;
  color: var(--c-focused-input-text);
}

.c-focused-input-dialog__description {
  margin: 0;
  max-width: 27ch;
  font-size: 14px;
  line-height: 1.7;
  color: var(--c-focused-input-muted);
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
  color: var(--c-focused-input-muted);
}

.c-focused-input-dialog__form :deep(input),
.c-focused-input-dialog__form :deep(textarea),
.c-focused-input-dialog__form :deep(select) {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid var(--c-focused-input-line);
  border-radius: 0.8rem;
  background: var(--c-focused-input-surface-muted);
  font-size: 14px;
  color: var(--c-focused-input-text);
  outline: none;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--c-focused-input-surface) 40%, transparent);
}

.c-focused-input-dialog__form :deep(input::placeholder),
.c-focused-input-dialog__form :deep(textarea::placeholder) {
  color: var(--c-focused-input-faint);
}

.c-focused-input-dialog__form :deep(input:focus),
.c-focused-input-dialog__form :deep(textarea:focus),
.c-focused-input-dialog__form :deep(select:focus) {
  border-color: color-mix(in srgb, var(--c-focused-input-accent) 32%, var(--c-focused-input-line));
  background: var(--c-focused-input-surface);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--c-focused-input-accent) 12%, transparent);
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
  color: var(--c-focused-input-muted);
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='ghost']:hover) {
  background: color-mix(in srgb, var(--c-focused-input-line) 18%, var(--c-focused-input-surface-muted));
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='primary']) {
  padding: 0.7rem 1.5rem;
  background: color-mix(in srgb, var(--c-focused-input-accent) 92%, var(--c-focused-input-text));
  color: var(--color-text-inverse);
  box-shadow: 0 12px 24px color-mix(in srgb, var(--c-focused-input-accent) 20%, transparent);
}

.c-focused-input-dialog__footer :deep([data-c-modal-action='primary']:hover) {
  background: color-mix(in srgb, var(--c-focused-input-accent) 82%, var(--c-focused-input-text));
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
