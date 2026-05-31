import { ref } from 'vue'

import {
  destroyInstance as apiDestroyInstance,
  extendInstance as apiExtendInstance,
  requestInstanceAccess,
} from '@/api/instance'
import type { InstanceExtendData } from '@/api/contracts'
import { useClipboard } from '@/shared/model/common/useClipboard'
import {
  confirmDestructiveAction,
  type DestructiveConfirmOptions,
} from '@/shared/model/common/useDestructiveConfirm'
import { useToast } from '@/shared/model/common/useToast'

interface InstanceWorkflowTarget {
  id: string
}

interface UseInstanceWorkflowActionsOptions<T extends InstanceWorkflowTarget> {
  resolveTarget: (id?: string) => T | null | undefined
  getExtendBlockedMessage?: (target: T) => string | null
  getDestroyBlockedMessage?: (target: T) => string | null
  onExtended: (context: { target: T; result: InstanceExtendData | null }) => Promise<void> | void
  onDestroyed: (context: { target: T }) => Promise<void> | void
  openErrorMessage?: string
  extendSuccessMessage?: string | null
  extendErrorMessage?: string
  destroySuccessMessage?: string | null
  destroyErrorMessage?: string
  destroyConfirmOptions?: DestructiveConfirmOptions
}

const defaultDestroyConfirmOptions: DestructiveConfirmOptions = {
  title: '确认销毁实例',
  message: '确定要销毁该实例吗？此操作不可恢复。',
  confirmButtonText: '确认销毁',
  cancelButtonText: '取消',
}

function resolveErrorMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.message.trim() ? error.message : fallback
}

export function useInstanceWorkflowActions<T extends InstanceWorkflowTarget>(
  options: UseInstanceWorkflowActionsOptions<T>
) {
  const toast = useToast()
  const { copy } = useClipboard()

  const opening = ref(false)
  const extending = ref(false)
  const destroying = ref(false)

  async function openInstance(id?: string) {
    const target = options.resolveTarget(id)
    const targetId = id ?? target?.id
    if (!targetId || opening.value) {
      return
    }

    opening.value = true
    try {
      const result = await requestInstanceAccess(targetId)
      const command = result.access?.protocol === 'tcp' ? result.access.command?.trim() ?? '' : ''
      if (command) {
        await copy(command)
        toast.info('TCP 连接命令已复制')
        return
      }
      if (typeof window !== 'undefined') {
        window.open(result.access_url, '_blank', 'noopener,noreferrer')
      }
    } catch (error) {
      toast.error(options.openErrorMessage ?? '打开目标失败')
    } finally {
      opening.value = false
    }
  }

  async function extendInstance(id?: string) {
    const target = options.resolveTarget(id)
    if (!target || extending.value) {
      return
    }

    const blockedMessage = options.getExtendBlockedMessage?.(target)
    if (blockedMessage) {
      toast.error(blockedMessage)
      return
    }

    extending.value = true
    try {
      const result = await apiExtendInstance(target.id)
      await options.onExtended({ target, result })
      if (options.extendSuccessMessage) {
        toast.success(options.extendSuccessMessage)
      }
    } catch (error) {
      toast.error(options.extendErrorMessage ?? '延时失败')
    } finally {
      extending.value = false
    }
  }

  async function destroyInstance(id?: string) {
    const target = options.resolveTarget(id)
    if (!target || destroying.value) {
      return
    }

    const blockedMessage = options.getDestroyBlockedMessage?.(target)
    if (blockedMessage) {
      toast.error(blockedMessage)
      return
    }

    const confirmed = await confirmDestructiveAction(
      options.destroyConfirmOptions ?? defaultDestroyConfirmOptions
    )
    if (!confirmed) {
      return
    }

    destroying.value = true
    try {
      await apiDestroyInstance(target.id)
      await options.onDestroyed({ target })
      if (options.destroySuccessMessage) {
        toast.success(options.destroySuccessMessage)
      }
    } catch (error) {
      toast.error(resolveErrorMessage(error, options.destroyErrorMessage ?? '销毁实例失败'))
    } finally {
      destroying.value = false
    }
  }

  return {
    opening,
    extending,
    destroying,
    openInstance,
    extendInstance,
    destroyInstance,
  }
}
