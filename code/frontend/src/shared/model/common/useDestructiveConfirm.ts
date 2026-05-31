import { computed, nextTick, ref } from 'vue'

export interface DestructiveConfirmOptions {
  message: string
  title?: string
  warning?: string
  confirmButtonText?: string
  cancelButtonText?: string
  closeOnBackdrop?: boolean
  closeOnEscape?: boolean
}

interface ActiveDestructiveConfirm {
  message: string
  title: string
  warning: string
  confirmButtonText: string
  cancelButtonText: string
  closeOnBackdrop: boolean
  closeOnEscape: boolean
}

const activeConfirm = ref<ActiveDestructiveConfirm | null>(null)

let resolveCurrentConfirm: ((value: boolean) => void) | null = null
let restoreFocusTarget: HTMLElement | null = null

function resolveDefaultWarning(options: DestructiveConfirmOptions): string {
  if (typeof options.warning === 'string') {
    return options.warning.trim()
  }

  if (/(不可恢复|不可撤销)/.test(options.message)) {
    return ''
  }

  const combined = `${options.title || ''} ${options.message} ${options.confirmButtonText || ''}`
  if (/(删除|销毁|移除)/.test(combined)) {
    return '此操作不可恢复，请确认后继续。'
  }

  if (/(结束|覆盖|踢出|强制)/.test(combined)) {
    return '确认后将立即生效，请确认后继续。'
  }

  return ''
}

function normalizeOptions(options: DestructiveConfirmOptions): ActiveDestructiveConfirm {
  return {
    message: options.message,
    title: options.title?.trim() || '确认操作',
    warning: resolveDefaultWarning(options),
    confirmButtonText: options.confirmButtonText?.trim() || '确认',
    cancelButtonText: options.cancelButtonText?.trim() || '取消',
    closeOnBackdrop: options.closeOnBackdrop ?? true,
    closeOnEscape: options.closeOnEscape ?? true,
  }
}

function restoreTriggerFocus(): void {
  const target = restoreFocusTarget
  restoreFocusTarget = null

  if (!target) {
    return
  }

  void nextTick(() => {
    if (!target.isConnected) {
      return
    }

    target.focus()
  })
}

function settleCurrentConfirm(result: boolean): void {
  const resolver = resolveCurrentConfirm
  resolveCurrentConfirm = null
  activeConfirm.value = null
  resolver?.(result)
  restoreTriggerFocus()
}

export async function confirmDestructiveAction({
  message,
  title,
  warning,
  confirmButtonText,
  cancelButtonText,
  closeOnBackdrop,
  closeOnEscape,
}: DestructiveConfirmOptions): Promise<boolean> {
  if (resolveCurrentConfirm) {
    // 全局确认框只能同时存在一个实例；后来的请求必须先把前一个 Promise 收口，
    // 否则调用方会拿到一个永远不 resolve 的悬空确认结果。
    settleCurrentConfirm(false)
  }

  restoreFocusTarget =
    typeof document !== 'undefined' && document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null

  activeConfirm.value = normalizeOptions({
    message,
    title,
    warning,
    confirmButtonText,
    cancelButtonText,
    closeOnBackdrop,
    closeOnEscape,
  })

  return new Promise((resolve) => {
    resolveCurrentConfirm = resolve
  })
}

export function useDestructiveConfirmState() {
  return {
    current: computed(() => activeConfirm.value),
    visible: computed(() => activeConfirm.value !== null),
    confirm: () => settleCurrentConfirm(true),
    cancel: () => settleCurrentConfirm(false),
  }
}
