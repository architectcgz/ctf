import { type Ref } from 'vue'

import {
  destroyInstance as apiDestroyInstance,
  extendInstance,
  requestInstanceAccess,
} from '@/api/instance'
import { useClipboard } from '@/composables/useClipboard'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

import type { InstanceViewModel } from './useInstanceListPage'

interface UseInstanceOperationsOptions {
  instances: Ref<InstanceViewModel[]>
  warnedInstances: Set<string>
  warningInstance: Ref<InstanceViewModel | null>
  showWarning: Ref<boolean>
  isInstanceManualActionAllowed: (instance: InstanceViewModel) => boolean
  isAWDTeamInstance: (instance: Pick<InstanceViewModel, 'contest_mode' | 'share_scope'>) => boolean
  calculateRemaining: (expiresAt: string) => number
  loadInstances: () => Promise<void>
}

export function useInstanceOperations(options: UseInstanceOperationsOptions) {
  const {
    instances,
    warnedInstances,
    warningInstance,
    showWarning,
    isInstanceManualActionAllowed,
    isAWDTeamInstance,
    calculateRemaining,
    loadInstances,
  } = options
  const toast = useToast()
  const { copy } = useClipboard()

  async function copyAddress(address: string) {
    if (!address) {
      return
    }
    await copy(address)
  }

  async function extendTime(id: string) {
    const target = instances.value.find((instance) => instance.id === id)
    if (target && !isInstanceManualActionAllowed(target)) {
      toast.error(
        isAWDTeamInstance(target) ? 'AWD 队伍实例不支持在此处延时或销毁' : '共享实例不支持手动延时'
      )
      return
    }
    try {
      const result = await extendInstance(id)
      if (result) {
        instances.value = instances.value.map((instance) =>
          instance.id === id
            ? {
                ...instance,
                remaining: calculateRemaining(result.expires_at),
                expires_at: result.expires_at,
                remaining_extends: result.remaining_extends,
              }
            : instance
        )
        warnedInstances.delete(id)
      } else {
        await loadInstances()
      }
    } catch (error) {
      console.error('延时失败:', error)
      toast.error('延时失败，请稍后重试')
    }
  }

  async function openTarget(id: string) {
    try {
      const result = await requestInstanceAccess(id)
      const command = result.access?.protocol === 'tcp' ? result.access.command : ''
      if (command) {
        await copy(command)
        toast.info('TCP 连接命令已复制')
        return
      }
      window.open(result.access_url, '_blank', 'noopener,noreferrer')
    } catch (error) {
      console.error('打开目标失败:', error)
      toast.error('打开目标失败，请稍后重试')
    }
  }

  async function destroyInstance(id: string) {
    const target = instances.value.find((instance) => instance.id === id)
    if (target && !isInstanceManualActionAllowed(target)) {
      toast.error(
        isAWDTeamInstance(target) ? 'AWD 队伍实例不支持在此处延时或销毁' : '共享实例不支持手动销毁'
      )
      return
    }
    const confirmed = await confirmDestructiveAction({
      title: '确认销毁实例',
      message: '确定要销毁该实例吗？此操作不可恢复。',
      confirmButtonText: '确认销毁',
      cancelButtonText: '取消',
    })
    if (!confirmed) {
      return
    }

    try {
      await apiDestroyInstance(id)
      instances.value = instances.value.filter((instance) => instance.id !== id)
      warnedInstances.delete(id)
      if (warningInstance.value?.id === id) {
        warningInstance.value = null
        showWarning.value = false
      }
    } catch (error) {
      console.error('销毁失败:', error)
      const message =
        error instanceof Error && error.message.trim() ? error.message : '销毁失败，请稍后重试'
      toast.error(message)
    }
  }

  return {
    copyAddress,
    extendTime,
    openTarget,
    destroyInstance,
  }
}
