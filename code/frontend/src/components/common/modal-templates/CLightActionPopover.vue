<script setup lang="ts">
import { MessageSquare, SendHorizonal } from 'lucide-vue-next'
import { computed, onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    description?: string
    width?: string
    closeOnOutside?: boolean
  }>(),
  {
    title: '发现题目问题？',
    description: '请简要描述您遇到的环境或描述错误。',
    width: '18rem',
    closeOnOutside: true,
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
  submit: []
}>()

const rootRef = ref<HTMLDivElement | null>(null)
const panelStyle = computed<Record<string, string>>(() => ({
  '--c-light-action-popover-width': props.width,
}))

function closePopover(): void {
  emit('update:modelValue', false)
  emit('close')
}

function togglePopover(): void {
  emit('update:modelValue', !props.modelValue)
}

function handleWindowClick(event: MouseEvent): void {
  if (!props.modelValue || !props.closeOnOutside) return
  const target = event.target
  if (!(target instanceof Node)) return
  if (rootRef.value?.contains(target)) return
  closePopover()
}

watch(
  () => props.modelValue,
  (open) => {
    window.removeEventListener('click', handleWindowClick)
    if (open && props.closeOnOutside) {
      window.addEventListener('click', handleWindowClick)
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  window.removeEventListener('click', handleWindowClick)
})
</script>

<template>
  <div ref="rootRef" class="c-light-action-popover relative inline-flex">
    <slot name="trigger" :open="props.modelValue" :toggle="togglePopover">
      <button
        type="button"
        class="c-light-action-popover__trigger"
        :class="{ 'c-light-action-popover__trigger--active': props.modelValue }"
        @click.stop="togglePopover"
      >
        <MessageSquare :size="16" />
        题目反馈
      </button>
    </slot>

    <div
      v-if="props.modelValue"
      class="c-light-action-popover__panel absolute left-1/2 top-full z-50 mt-3 -translate-x-1/2"
      :style="panelStyle"
    >
      <div class="c-light-action-popover__body">
        <h3 class="c-light-action-popover__title">{{ props.title }}</h3>
        <p class="c-light-action-popover__description">{{ props.description }}</p>

        <slot>
          <textarea
            class="c-light-action-popover__textarea"
            placeholder="例如：实例无法连接，或附件下载 404..."
          />
        </slot>

        <div class="c-light-action-popover__actions">
          <slot name="actions" :close="closePopover">
            <button
              type="button"
              class="c-light-action-popover__action c-light-action-popover__action--ghost"
              @click="closePopover"
            >
              取消
            </button>
            <button
              type="button"
              class="c-light-action-popover__action c-light-action-popover__action--primary"
              @click="emit('submit')"
            >
              发送反馈
              <SendHorizonal :size="12" />
            </button>
          </slot>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.c-light-action-popover {
  --c-light-action-popover-surface: color-mix(in srgb, var(--color-bg-elevated) 96%, var(--color-bg-surface));
  --c-light-action-popover-surface-muted: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --c-light-action-popover-line: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --c-light-action-popover-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --c-light-action-popover-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --c-light-action-popover-muted: color-mix(in srgb, var(--color-text-secondary) 92%, transparent);
  --c-light-action-popover-faint: color-mix(in srgb, var(--color-text-muted) 94%, transparent);
  --c-light-action-popover-accent: var(--color-primary);
}

.c-light-action-popover__trigger {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 0.75rem;
  border: 1px solid var(--c-light-action-popover-line);
  background: var(--c-light-action-popover-surface);
  padding: 0.625rem 1.25rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--c-light-action-popover-muted);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.c-light-action-popover__trigger:hover,
.c-light-action-popover__trigger--active {
  border-color: color-mix(in srgb, var(--c-light-action-popover-accent) 18%, var(--c-light-action-popover-line-strong));
  background: color-mix(in srgb, var(--c-light-action-popover-accent) 8%, var(--c-light-action-popover-surface));
  color: var(--c-light-action-popover-text);
}

.c-light-action-popover__panel {
  width: min(var(--c-light-action-popover-width, 18rem), calc(100vw - 2rem));
  overflow: hidden;
  border-radius: 1rem;
  border: 1px solid var(--c-light-action-popover-line-strong);
  background: var(--c-light-action-popover-surface);
  box-shadow:
    0 18px 40px color-mix(in srgb, var(--color-shadow-strong) 14%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.c-light-action-popover__body {
  padding: 1.25rem;
}

.c-light-action-popover__title {
  margin: 0 0 0.25rem;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--c-light-action-popover-text);
}

.c-light-action-popover__description {
  margin: 0 0 1rem;
  font-size: 0.75rem;
  line-height: 1.5;
  color: var(--c-light-action-popover-muted);
}

.c-light-action-popover__textarea {
  margin-bottom: 0.75rem;
  min-height: 5rem;
  width: 100%;
  resize: none;
  border-radius: 0.75rem;
  border: 1px solid var(--c-light-action-popover-line);
  background: var(--c-light-action-popover-surface-muted);
  padding: 0.5rem 0.75rem;
  font-size: 0.8125rem;
  color: var(--c-light-action-popover-text);
  outline: none;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background-color 0.2s ease;
}

.c-light-action-popover__textarea::placeholder {
  color: var(--c-light-action-popover-faint);
}

.c-light-action-popover__textarea:focus {
  border-color: color-mix(in srgb, var(--c-light-action-popover-accent) 30%, var(--c-light-action-popover-line));
  background: var(--c-light-action-popover-surface);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--c-light-action-popover-accent) 14%, transparent);
}

.c-light-action-popover__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.c-light-action-popover__action {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 0.75rem;
  padding: 0.5rem 1rem;
  font-size: 0.75rem;
  font-weight: 600;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease;
}

.c-light-action-popover__action--ghost {
  color: var(--c-light-action-popover-muted);
}

.c-light-action-popover__action--ghost:hover {
  color: var(--c-light-action-popover-text);
}

.c-light-action-popover__action--primary {
  border: 1px solid color-mix(in srgb, var(--c-light-action-popover-accent) 22%, var(--c-light-action-popover-line));
  background: color-mix(in srgb, var(--c-light-action-popover-accent) 92%, var(--c-light-action-popover-text));
  color: var(--color-text-inverse);
}

.c-light-action-popover__action--primary:hover {
  border-color: color-mix(in srgb, var(--c-light-action-popover-accent) 30%, var(--c-light-action-popover-line-strong));
  background: color-mix(in srgb, var(--c-light-action-popover-accent) 82%, var(--c-light-action-popover-text));
}

.c-light-action-popover__trigger:focus-visible,
.c-light-action-popover__textarea:focus-visible,
.c-light-action-popover__action:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--c-light-action-popover-accent) 42%, white);
  outline-offset: 2px;
}
</style>
