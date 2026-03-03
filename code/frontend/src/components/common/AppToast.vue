<template>
  <div class="fixed top-4 right-4 z-50 flex w-[360px] flex-col gap-2 max-w-[calc(100vw-2rem)]">
    <div
      v-for="item in toasts"
      :key="item.id"
      class="group flex items-start justify-between gap-3 rounded-lg border border-border bg-surface/95 px-4 py-3 shadow-lg backdrop-blur"
      :class="toneClass(item.type)"
      role="status"
    >
      <div class="min-w-0">
        <div class="text-sm font-semibold leading-5">{{ title(item.type) }}</div>
        <div class="mt-0.5 text-sm text-text-secondary break-words">{{ item.message }}</div>
      </div>

      <button
        type="button"
        class="shrink-0 rounded-md px-2 py-1 text-xs text-text-muted hover:bg-elevated hover:text-text-primary"
        @click="toast.dismiss(item.id)"
      >
        关闭
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { provideToast, type ToastType, useToast, useToastState } from '@/composables/useToast'

provideToast()

const toast = useToast()
const { toasts } = useToastState()

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

function toneClass(type: ToastType): string {
  switch (type) {
    case 'success':
      return 'border-emerald-500/20'
    case 'warning':
      return 'border-amber-500/20'
    case 'info':
      return 'border-cyan-500/20'
    case 'error':
      return 'border-red-500/20'
  }
}
</script>

