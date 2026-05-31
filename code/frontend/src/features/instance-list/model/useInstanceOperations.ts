import { type Ref } from 'vue'

import { useInstanceWorkflowActions } from '@/features/instance-workflow'
import { useClipboard } from '@/shared/model/common/useClipboard'

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
  const { copy } = useClipboard()
  const {
    extendInstance: extendWorkflowInstance,
    openInstance: openWorkflowInstance,
    destroyInstance: destroyWorkflowInstance,
  } = useInstanceWorkflowActions({
    resolveTarget: (id) =>
      instances.value.find((instance) => instance.id === id) ??
      (id ? ({ id } as InstanceViewModel) : null),
    getExtendBlockedMessage: (target) =>
      !isInstanceManualActionAllowed(target)
        ? isAWDTeamInstance(target)
          ? 'AWD 队伍实例不支持在此处延时或销毁'
          : '共享实例不支持手动延时'
        : null,
    getDestroyBlockedMessage: (target) =>
      !isInstanceManualActionAllowed(target)
        ? isAWDTeamInstance(target)
          ? 'AWD 队伍实例不支持在此处延时或销毁'
          : '共享实例不支持手动销毁'
        : null,
    onExtended: async ({ target, result }) => {
      if (result) {
        instances.value = instances.value.map((instance) =>
          instance.id === target.id
            ? {
                ...instance,
                remaining: calculateRemaining(result.expires_at),
                expires_at: result.expires_at,
                remaining_extends: result.remaining_extends,
              }
            : instance
        )
        warnedInstances.delete(target.id)
        return
      }

      await loadInstances()
    },
    onDestroyed: ({ target }) => {
      instances.value = instances.value.filter((instance) => instance.id !== target.id)
      warnedInstances.delete(target.id)
      if (warningInstance.value?.id === target.id) {
        warningInstance.value = null
        showWarning.value = false
      }
    },
    openErrorMessage: '打开目标失败，请稍后重试',
    extendErrorMessage: '延时失败，请稍后重试',
    destroyErrorMessage: '销毁失败，请稍后重试',
  })

  async function copyAddress(address: string) {
    if (!address) {
      return
    }
    await copy(address)
  }

  async function extendTime(id: string) {
    await extendWorkflowInstance(id)
  }

  async function openTarget(id: string) {
    await openWorkflowInstance(id)
  }

  async function destroyInstance(id: string) {
    await destroyWorkflowInstance(id)
  }

  return {
    copyAddress,
    extendTime,
    openTarget,
    destroyInstance,
  }
}
