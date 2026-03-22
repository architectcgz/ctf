<template>
  <div class="fixed right-4 top-4 z-50 flex max-w-[calc(100vw-2rem)] w-[380px] flex-col gap-3">
    <div
      v-for="item in toasts"
      :key="item.id"
      class="group relative overflow-hidden rounded-[22px] border border-border bg-surface px-4 py-3.5 shadow-[0_18px_40px_var(--color-shadow-soft)]"
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
            <component :is="toneMeta(item.type).icon" class="h-5 w-5" :style="{ color: toneMeta(item.type).accentColor }" />
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
          class="shrink-0 rounded-xl border border-border bg-base/70 px-2.5 py-1.5 text-xs font-medium text-text-secondary transition-colors hover:text-text-primary"
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
      boxShadow: '0 18px 40px var(--color-shadow-soft)',
      borderColor: 'color-mix(in srgb, var(--color-success) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'rgba(63, 185, 80, 0.12)',
      borderColor: 'rgba(63, 185, 80, 0.28)',
    },
    closeStyle: {
      borderColor: 'rgba(63, 185, 80, 0.18)',
      backgroundColor: 'rgba(63, 185, 80, 0.08)',
    },
  },
  warning: {
    icon: AlertTriangle,
    accentColor: 'var(--color-warning)',
    containerStyle: {
      boxShadow: '0 18px 40px var(--color-shadow-soft)',
      borderColor: 'color-mix(in srgb, var(--color-warning) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'rgba(210, 153, 34, 0.12)',
      borderColor: 'rgba(210, 153, 34, 0.28)',
    },
    closeStyle: {
      borderColor: 'rgba(210, 153, 34, 0.18)',
      backgroundColor: 'rgba(210, 153, 34, 0.08)',
    },
  },
  info: {
    icon: Info,
    accentColor: 'var(--color-primary)',
    containerStyle: {
      boxShadow: '0 18px 40px var(--color-shadow-soft)',
      borderColor: 'color-mix(in srgb, var(--color-primary) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'var(--color-primary-soft)',
      borderColor: 'color-mix(in srgb, var(--color-primary) 28%, transparent)',
    },
    closeStyle: {
      borderColor: 'color-mix(in srgb, var(--color-primary) 18%, transparent)',
      backgroundColor: 'rgba(8, 145, 178, 0.08)',
    },
  },
  error: {
    icon: OctagonX,
    accentColor: 'var(--color-danger)',
    containerStyle: {
      boxShadow: '0 18px 40px var(--color-shadow-soft)',
      borderColor: 'color-mix(in srgb, var(--color-danger) 24%, var(--color-border-default))',
    },
    iconWrapStyle: {
      backgroundColor: 'rgba(248, 81, 73, 0.12)',
      borderColor: 'rgba(248, 81, 73, 0.28)',
    },
    closeStyle: {
      borderColor: 'rgba(248, 81, 73, 0.18)',
      backgroundColor: 'rgba(248, 81, 73, 0.08)',
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
