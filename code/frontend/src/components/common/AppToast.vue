<template>
  <div class="fixed right-4 top-4 z-50 flex max-w-[calc(100vw-2rem)] w-[380px] flex-col gap-3">
    <div
      v-for="item in toasts"
      :key="item.id"
      :class="[
        'app-toast-item group relative overflow-hidden rounded-[22px] border px-4 py-3.5',
      ]"
      :style="toneMeta(item.type).containerStyle"
      :role="item.type === 'error' ? 'alert' : 'status'"
      :aria-live="item.type === 'error' ? 'assertive' : 'polite'"
    >
      <div
        class="pointer-events-none absolute inset-y-0 left-0 w-1.5"
        :style="{ backgroundColor: toneMeta(item.type).accentColor }"
      />

      <div class="flex items-start justify-between gap-3 pl-2">
        <div class="flex min-w-0 items-start gap-3">
          <div
            class="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl border"
            :style="toneMeta(item.type).iconWrapStyle"
          >
            <component
              :is="toneMeta(item.type).icon"
              class="h-5 w-5"
              :style="{ color: toneMeta(item.type).accentColor }"
            />
          </div>

          <div class="min-w-0">
            <div class="text-sm font-semibold leading-5 text-text-primary">
              {{ title(item.type) }}
            </div>
            <div class="mt-1 break-words text-sm leading-6 text-text-primary/92">
              {{ item.message }}
            </div>
          </div>
        </div>

        <button
          type="button"
          class="app-toast-close shrink-0 rounded-xl border px-2.5 py-1.5 text-xs font-medium transition-colors hover:text-text-primary"
          :style="toneMeta(item.type).closeStyle"
          @click="toast.dismiss(item.id)"
        >
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { AlertTriangle, CheckCircle2, Info, OctagonX } from 'lucide-vue-next'

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
    accentColor: 'var(--color-success)',
    containerStyle: {
      borderColor: 'color-mix(in srgb, var(--color-success) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'color-mix(in srgb, var(--color-success) 12%, var(--color-bg-surface))',
      borderColor: 'color-mix(in srgb, var(--color-success) 30%, var(--color-border-default))',
    },
    closeStyle: {
      borderColor: 'color-mix(in srgb, var(--color-success) 22%, var(--color-border-default))',
      backgroundColor: 'color-mix(in srgb, var(--color-success) 8%, var(--color-bg-surface))',
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
</script>

<style scoped>
.app-toast-item {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base)),
    color-mix(in srgb, var(--color-bg-surface) 86%, var(--color-bg-base))
  );
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.app-toast-close {
  border-color: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  background: color-mix(in srgb, var(--color-bg-base) 76%, var(--color-bg-surface));
  color: var(--color-text-secondary);
}
</style>
