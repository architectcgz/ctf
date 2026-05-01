<template>
  <div class="app-toast-stack">
    <div
      v-for="item in toasts"
      :key="item.id"
      class="app-toast-item"
      :style="toastStyle(item.type)"
      :role="item.type === 'error' ? 'alert' : 'status'"
      :aria-live="item.type === 'error' ? 'assertive' : 'polite'"
    >
      <div class="app-toast-content">
        <div
          class="app-toast-icon"
          :style="toneMeta(item.type).iconWrapStyle"
        >
          <component
            :is="toneMeta(item.type).icon"
            class="app-toast-type-icon"
            :style="{ color: toneMeta(item.type).accentColor }"
          />
        </div>

        <div class="app-toast-copy">
          <div class="app-toast-title">
            {{ title(item.type) }}
          </div>
          <div class="app-toast-message">
            {{ item.message }}
          </div>
        </div>

        <button
          type="button"
          class="app-toast-close"
          :style="toneMeta(item.type).closeStyle"
          aria-label="关闭提示"
          title="关闭提示"
          @click="toast.dismiss(item.id)"
        >
          <X class="app-toast-close-icon" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { AlertTriangle, CheckCircle2, Info, OctagonX, X } from 'lucide-vue-next'

import { provideToast, type ToastType, useToast, useToastState } from '@/composables/useToast'

provideToast()

const toast = useToast()
const { toasts } = useToastState()

interface ToastToneMeta {
  icon: Component
  accentColor: string
  containerStyle: Record<string, string>
  iconWrapStyle: Record<string, string>
  closeStyle: Record<string, string>
}

const toneMap: Record<ToastType, ToastToneMeta> = {
  success: {
    icon: CheckCircle2,
    accentColor: 'var(--color-primary)',
    containerStyle: {
      borderColor: 'color-mix(in srgb, var(--color-primary) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'color-mix(in srgb, var(--color-primary) 12%, var(--color-bg-surface))',
      borderColor: 'color-mix(in srgb, var(--color-primary) 30%, var(--color-border-default))',
    },
    closeStyle: {
      borderColor: 'color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default))',
      backgroundColor: 'color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface))',
    },
  },
  warning: {
    icon: AlertTriangle,
    accentColor: 'var(--color-warning)',
    containerStyle: {
      borderColor: 'color-mix(in srgb, var(--color-warning) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'color-mix(in srgb, var(--color-warning) 12%, var(--color-bg-surface))',
      borderColor: 'color-mix(in srgb, var(--color-warning) 30%, var(--color-border-default))',
    },
    closeStyle: {
      borderColor: 'color-mix(in srgb, var(--color-warning) 22%, var(--color-border-default))',
      backgroundColor: 'color-mix(in srgb, var(--color-warning) 8%, var(--color-bg-surface))',
    },
  },
  info: {
    icon: Info,
    accentColor: 'var(--color-primary)',
    containerStyle: {
      borderColor: 'color-mix(in srgb, var(--color-primary) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'color-mix(in srgb, var(--color-primary) 12%, var(--color-bg-surface))',
      borderColor: 'color-mix(in srgb, var(--color-primary) 30%, var(--color-border-default))',
    },
    closeStyle: {
      borderColor: 'color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default))',
      backgroundColor: 'color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface))',
    },
  },
  error: {
    icon: OctagonX,
    accentColor: 'var(--color-danger)',
    containerStyle: {
      borderColor: 'color-mix(in srgb, var(--color-danger) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'color-mix(in srgb, var(--color-danger) 12%, var(--color-bg-surface))',
      borderColor: 'color-mix(in srgb, var(--color-danger) 30%, var(--color-border-default))',
    },
    closeStyle: {
      borderColor: 'color-mix(in srgb, var(--color-danger) 22%, var(--color-border-default))',
      backgroundColor: 'color-mix(in srgb, var(--color-danger) 8%, var(--color-bg-surface))',
    },
  },
}

function title(type: ToastType): string {
  switch (type) {
    case 'success':
      return '成功'
    case 'warning':
      return '提示'
    case 'info':
      return '信息'
    case 'error':
      return '错误'
  }
}

function toneMeta(type: ToastType): ToastToneMeta {
  return toneMap[type]
}

function toastStyle(type: ToastType): Record<string, string> {
  const meta = toneMeta(type)

  return {
    ...meta.containerStyle,
    '--app-toast-accent-color': meta.accentColor,
  }
}
</script>

<style scoped>
.app-toast-stack {
  position: fixed;
  top: var(--space-4);
  right: var(--space-4);
  z-index: 50;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  width: min(24rem, calc(100vw - (var(--space-4) * 2)));
  pointer-events: none;
}

.app-toast-item {
  position: relative;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  border-radius: var(--ui-dialog-radius);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-bg-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base))
  );
  box-shadow: 0 var(--space-4-5) var(--space-10) var(--color-shadow-soft);
  pointer-events: auto;
}

.app-toast-item::before {
  position: absolute;
  inset-block: 0;
  inset-inline-start: 0;
  width: 38%;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--app-toast-accent-color) 10%, transparent),
    transparent 82%
  );
  content: '';
  pointer-events: none;
}

.app-toast-content {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3-5) var(--space-3-5) var(--space-3-5) var(--space-4-5);
}

.app-toast-icon {
  display: flex;
  width: var(--ui-control-height-sm);
  height: var(--ui-control-height-sm);
  flex-shrink: 0;
  align-self: center;
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 80%, transparent);
  border-radius: var(--ui-control-radius-md);
}

.app-toast-type-icon {
  width: var(--space-4-5);
  height: var(--space-4-5);
}

.app-toast-copy {
  min-width: 0;
}

.app-toast-title {
  font-size: var(--font-size-14);
  font-weight: 700;
  line-height: 1.35;
  color: var(--color-text-primary);
}

.app-toast-message {
  margin-top: var(--space-1);
  max-height: min(8rem, calc(100vh - var(--space-12)));
  overflow: auto;
  overflow-wrap: anywhere;
  font-size: var(--font-size-14);
  line-height: 1.6;
  color: color-mix(in srgb, var(--color-text-primary) 90%, var(--color-text-secondary));
}

.app-toast-close {
  display: inline-flex;
  width: var(--ui-control-height-sm);
  height: var(--ui-control-height-sm);
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  border-radius: var(--ui-control-radius-md);
  background: color-mix(in srgb, var(--color-bg-base) 76%, var(--color-bg-surface));
  color: var(--color-text-secondary);
  transition:
    border-color var(--ui-motion-fast),
    background-color var(--ui-motion-fast),
    color var(--ui-motion-fast);
}

.app-toast-close:hover {
  color: var(--color-text-primary);
}

.app-toast-close:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 54%, transparent);
  outline-offset: var(--space-0-5);
}

.app-toast-close-icon {
  width: var(--space-4);
  height: var(--space-4);
}

@media (max-width: 40rem) {
  .app-toast-stack {
    inset-inline: var(--space-3);
    top: var(--space-3);
    width: auto;
  }

  .app-toast-content {
    padding-inline-end: var(--space-3);
  }
}
</style>
