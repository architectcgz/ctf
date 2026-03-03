import { computed, inject, provide, ref } from 'vue'

import { TOAST_DURATION } from '@/utils/constants'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface ToastItem {
  id: string
  type: ToastType
  message: string
  createdAt: number
  durationMs: number
}

export interface ToastApi {
  success: (message: string) => void
  error: (message: string) => void
  warning: (message: string) => void
  info: (message: string) => void
  dismiss: (id: string) => void
}

const TOAST_KEY: unique symbol = Symbol('toast')

const toasts = ref<ToastItem[]>([])

function add(type: ToastType, message: string, durationMs: number): void {
  const id = `${Date.now()}-${Math.random().toString(16).slice(2)}`
  const item: ToastItem = { id, type, message, createdAt: Date.now(), durationMs }
  toasts.value = [...toasts.value, item]
  window.setTimeout(() => dismiss(id), durationMs)
}

function dismiss(id: string): void {
  toasts.value = toasts.value.filter((t) => t.id !== id)
}

const fallbackToast: ToastApi = {
  success: (message) => add('success', message, TOAST_DURATION.SUCCESS),
  info: (message) => add('info', message, TOAST_DURATION.INFO),
  warning: (message) => add('warning', message, TOAST_DURATION.WARNING),
  error: (message) => add('error', message, TOAST_DURATION.ERROR),
  dismiss,
}

export function provideToast(): ToastApi {
  provide(TOAST_KEY, fallbackToast)
  return fallbackToast
}

export function useToast(): ToastApi {
  return inject(TOAST_KEY, fallbackToast)
}

export function useToastState() {
  return {
    toasts: computed(() => toasts.value),
  }
}

