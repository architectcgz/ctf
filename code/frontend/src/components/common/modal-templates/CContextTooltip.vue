<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(
  defineProps<{
    title?: string
    description?: string
    width?: string
  }>(),
  {
    title: 'TLS 握手 (TLS Handshake)',
    description:
      '一种安全协议过程，用于在客户端和服务器之间建立加密通信连接。通常包含 ClientHello 和 ServerHello。',
    width: '16rem',
  }
)

const open = ref(false)
</script>

<template>
  <span
    class="c-context-tooltip c-context-tooltip__trigger relative inline-block cursor-help font-bold"
    @mouseenter="open = true"
    @mouseleave="open = false"
    @focusin="open = true"
    @focusout="open = false"
  >
    <slot>TLS 握手</slot>

    <div
      v-if="open"
      class="c-context-tooltip__panel absolute bottom-full left-1/2 z-50 mb-3 -translate-x-1/2 text-left font-normal leading-relaxed"
      :style="{ width: props.width }"
    >
      <div class="c-context-tooltip__title">{{ props.title }}</div>
      <div class="c-context-tooltip__description">{{ props.description }}</div>
      <div class="c-context-tooltip__arrow absolute left-1/2 top-full -translate-x-1/2" />
    </div>
  </span>
</template>

<style scoped>
.c-context-tooltip {
  --c-context-tooltip-surface: color-mix(in srgb, var(--color-bg-elevated) 18%, var(--color-bg-base));
  --c-context-tooltip-line: color-mix(in srgb, var(--color-border-default) 68%, transparent);
  --c-context-tooltip-text: color-mix(in srgb, var(--color-text-primary) 96%, white);
  --c-context-tooltip-muted: color-mix(in srgb, var(--color-text-secondary) 88%, white);
}

.c-context-tooltip__trigger {
  border-bottom: 1px dashed var(--c-context-tooltip-line);
  color: var(--color-text-primary);
}

.c-context-tooltip__panel {
  border-radius: 0.375rem;
  background: var(--c-context-tooltip-surface);
  padding: 1rem;
  font-size: 0.8125rem;
  color: var(--c-context-tooltip-text);
  box-shadow: 0 16px 36px color-mix(in srgb, var(--color-shadow-strong) 26%, transparent);
}

.c-context-tooltip__title {
  margin-bottom: 0.25rem;
  font-weight: 700;
  color: var(--c-context-tooltip-text);
}

.c-context-tooltip__description {
  color: var(--c-context-tooltip-muted);
}

.c-context-tooltip__arrow {
  border: 4px solid transparent;
  border-top-color: var(--c-context-tooltip-surface);
}
</style>
