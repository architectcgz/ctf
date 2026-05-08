<script setup lang="ts">
import { computed } from 'vue'
import { Trash2, X } from 'lucide-vue-next'

import ModalTemplateShell from '@/components/common/modal-templates/ModalTemplateShell.vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    username?: string
    targetName?: string
    title?: string
    description?: string
    warning?: string
    loading?: boolean
    confirmText?: string
    cancelText?: string
    loadingText?: string
    width?: string
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
  }>(),
  {
    username: '',
    targetName: '',
    title: '确认删除',
    description: '',
    warning: '此操作不可恢复，请确认后继续。',
    loading: false,
    confirmText: '确认删除',
    cancelText: '取消',
    loadingText: '',
    width: '27.5rem',
    closeOnBackdrop: true,
    closeOnEscape: true,
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: []
  cancel: []
}>()

const panelStyle = computed<Record<string, string>>(() => ({
  '--delete-confirm-modal-width': props.width,
}))

const resolvedTargetName = computed(() => props.targetName.trim() || props.username.trim())

const resolvedTitle = computed(() => props.title.trim() || '确认删除')

const resolvedDescription = computed(() => {
  if (props.description.trim()) {
    return props.description.trim()
  }

  if (resolvedTargetName.value) {
    return `确定要删除 ${resolvedTargetName.value} 吗？`
  }

  return '确定要执行当前危险操作吗？'
})

const resolvedWarning = computed(() => props.warning.trim())

const resolvedLoadingText = computed(() => {
  if (props.loadingText.trim()) {
    return props.loadingText.trim()
  }

  if (props.confirmText.includes('删除')) {
    return '删除中...'
  }

  return '处理中...'
})

function handleCancel(): void {
  if (props.loading) {
    return
  }

  emit('update:modelValue', false)
  emit('cancel')
}

function handleConfirm(): void {
  if (props.loading) {
    return
  }

  emit('confirm')
}

function handleShellOpenChange(next: boolean): void {
  if (!next) {
    handleCancel()
  }
}
</script>

<template>
  <ModalTemplateShell
    :open="modelValue"
    :panel-style="panelStyle"
    panel-class="delete-confirm-modal__panel"
    overlay-class="delete-confirm-modal__overlay"
    :aria-label="resolvedTitle"
    role="alertdialog"
    :close-on-backdrop="closeOnBackdrop && !loading"
    :close-on-escape="closeOnEscape && !loading"
    frosted
    @update:open="handleShellOpenChange"
  >
    <section class="delete-confirm-modal">
      <button
        type="button"
        class="delete-confirm-modal__close"
        :disabled="loading"
        aria-label="关闭确认弹窗"
        @click="handleCancel"
      >
        <X class="h-4 w-4" />
      </button>

      <div class="delete-confirm-modal__decoration">
        <div class="delete-confirm-modal__danger-icon">
          <Trash2 class="h-7 w-7" />
        </div>
      </div>

      <h2 class="delete-confirm-modal__title">
        {{ resolvedTitle }}
      </h2>

      <p class="delete-confirm-modal__description">
        {{ resolvedDescription }}
      </p>

      <p
        v-if="resolvedWarning"
        class="delete-confirm-modal__warning"
      >
        {{ resolvedWarning }}
      </p>

      <div class="delete-confirm-modal__actions">
        <button
          type="button"
          class="delete-confirm-modal__action delete-confirm-modal__action--cancel"
          :disabled="loading"
          @click="handleCancel"
        >
          {{ cancelText }}
        </button>

        <button
          type="button"
          class="delete-confirm-modal__action delete-confirm-modal__action--confirm"
          :disabled="loading"
          @click="handleConfirm"
        >
          <Trash2 class="h-4 w-4" />
          {{ loading ? resolvedLoadingText : confirmText }}
        </button>
      </div>
    </section>
  </ModalTemplateShell>
</template>

<style scoped>
.delete-confirm-modal__overlay {
  --modal-template-shell-overlay: color-mix(in srgb, var(--color-bg-base) 54%, transparent);
  --modal-shell-blur: var(--space-3);
}

.delete-confirm-modal__panel {
  width: min(var(--delete-confirm-modal-width, 27.5rem), 100%);
  max-height: calc(100vh - (var(--space-4) * 2));
  max-height: calc(100dvh - (var(--space-4) * 2));
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--color-danger) 18%, var(--color-border-default));
  border-radius: var(--ui-dialog-radius-wide);
  background:
    radial-gradient(
      circle at 50% 0,
      color-mix(in srgb, var(--color-danger) 12%, transparent),
      transparent 40%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base))
    );
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--color-bg-surface) 92%, white),
    var(--ui-dialog-shadow);
}

.delete-confirm-modal {
  position: relative;
  display: flex;
  max-height: inherit;
  flex-direction: column;
  gap: var(--space-4);
  overflow-y: auto;
  padding: var(--space-8) var(--space-7) var(--space-7);
  text-align: center;
}

.delete-confirm-modal::before {
  position: absolute;
  inset-inline: 0;
  top: var(--space-12);
  height: 5.5rem;
  background:
    radial-gradient(
      circle at 18% 78%,
      color-mix(in srgb, var(--color-danger) 14%, transparent) 0 1.75rem,
      transparent 1.8rem
    ),
    radial-gradient(
      circle at 82% 78%,
      color-mix(in srgb, var(--color-danger) 12%, transparent) 0 2.1rem,
      transparent 2.15rem
    );
  content: '';
  opacity: 0.88;
  pointer-events: none;
}

.delete-confirm-modal__close {
  position: absolute;
  top: var(--space-4);
  right: var(--space-4);
  z-index: 1;
  display: inline-flex;
  width: var(--ui-control-height-sm);
  height: var(--ui-control-height-sm);
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 88%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  color: var(--color-text-muted);
  transition:
    border-color var(--ui-motion-fast),
    background-color var(--ui-motion-fast),
    color var(--ui-motion-fast);
}

.delete-confirm-modal__close:hover:not(:disabled),
.delete-confirm-modal__close:focus-visible {
  border-color: color-mix(in srgb, var(--color-danger) 24%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-danger) 8%, var(--color-bg-surface));
  color: color-mix(in srgb, var(--color-danger) 92%, var(--color-text-primary));
}

.delete-confirm-modal__close:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.delete-confirm-modal__decoration,
.delete-confirm-modal__title,
.delete-confirm-modal__description,
.delete-confirm-modal__warning,
.delete-confirm-modal__actions {
  position: relative;
  z-index: 1;
}

.delete-confirm-modal__decoration {
  display: flex;
  justify-content: center;
}

.delete-confirm-modal__danger-icon {
  display: inline-flex;
  width: 4.5rem;
  height: 4.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--color-danger) 84%, white),
    color-mix(in srgb, var(--color-danger) 94%, var(--color-text-primary))
  );
  color: var(--ui-dialog-surface);
  box-shadow:
    0 var(--space-4) var(--space-8) color-mix(in srgb, var(--color-danger) 26%, transparent),
    0 0 0 var(--space-2-5) color-mix(in srgb, var(--color-danger) 10%, transparent);
}

.delete-confirm-modal__title {
  margin: 0;
  font-size: var(--font-size-22);
  font-weight: 700;
  line-height: 1.2;
  color: var(--color-text-primary);
}

.delete-confirm-modal__description {
  margin: 0;
  font-size: var(--font-size-15);
  line-height: 1.8;
  color: var(--color-text-secondary);
}

.delete-confirm-modal__warning {
  margin: calc(var(--space-1) * -1) 0 0;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--color-text-muted);
}

.delete-confirm-modal__actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3-5);
}

.delete-confirm-modal__action {
  display: inline-flex;
  min-height: var(--ui-control-height-md);
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  border-radius: calc(var(--ui-control-radius-md) + var(--space-0-5));
  padding-inline: var(--space-4);
  font-size: var(--font-size-15);
  font-weight: 600;
  transition:
    transform var(--ui-motion-fast),
    border-color var(--ui-motion-fast),
    background-color var(--ui-motion-fast),
    color var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast);
}

.delete-confirm-modal__action:disabled {
  opacity: 0.68;
  cursor: not-allowed;
  transform: none;
}

.delete-confirm-modal__action--cancel {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 90%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base));
  color: var(--color-text-primary);
}

.delete-confirm-modal__action--cancel:hover:not(:disabled),
.delete-confirm-modal__action--cancel:focus-visible {
  border-color: color-mix(in srgb, var(--color-text-secondary) 42%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-surface));
}

.delete-confirm-modal__action--confirm {
  border: 1px solid color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--color-danger) 88%, white),
    color-mix(in srgb, var(--color-danger) 96%, var(--color-text-primary))
  );
  color: var(--ui-dialog-surface);
  box-shadow: 0 var(--space-3) var(--space-6) color-mix(in srgb, var(--color-danger) 24%, transparent);
}

.delete-confirm-modal__action--confirm:hover:not(:disabled),
.delete-confirm-modal__action--confirm:focus-visible {
  transform: translateY(calc(var(--space-0-5) * -1));
  box-shadow: 0 var(--space-4) var(--space-7) color-mix(in srgb, var(--color-danger) 30%, transparent);
}

@media (max-width: 40rem) {
  .delete-confirm-modal {
    padding: var(--space-7) var(--space-5) var(--space-5);
  }

  .delete-confirm-modal__actions {
    grid-template-columns: 1fr;
  }
}
</style>
