import { ref, type Ref } from 'vue'

import type { ContestMode, InstanceSharing, InstanceStatus } from '@/api/contracts'

interface InstanceWarningBase {
  id: string
  status: InstanceStatus
  expires_at: string
  contest_mode?: ContestMode
  share_scope?: InstanceSharing
  remaining: number
}

interface UseInstanceWarningStateOptions<T extends InstanceWarningBase> {
  instances: Ref<T[]>
  warnedInstances: Set<string>
  warningThresholdSeconds: number
  canManualAction: (instance: T) => boolean
  onExtendInstance: (id: string) => Promise<void>
}

export function useInstanceWarningState<T extends InstanceWarningBase>(
  options: UseInstanceWarningStateOptions<T>
) {
  const {
    instances,
    warnedInstances,
    warningThresholdSeconds,
    canManualAction,
    onExtendInstance,
  } = options

  const showWarning = ref(false)
  const warningInstance = ref<T | null>(null)

  function updateCountdown() {
    const now = Date.now()

    instances.value = instances.value.map((instance) => {
      if (instance.status !== 'running') {
        return instance
      }
      if (!canManualAction(instance)) {
        return instance
      }

      const remaining = Math.max(
        0,
        Math.floor((new Date(instance.expires_at).getTime() - now) / 1000)
      )
      const next = {
        ...instance,
        remaining,
      }

      if (remaining < warningThresholdSeconds && !warnedInstances.has(instance.id)) {
        warnedInstances.add(instance.id)
        warningInstance.value = next
        showWarning.value = true
      }

      return next
    })
  }

  async function extendFromWarning() {
    if (warningInstance.value) {
      await onExtendInstance(warningInstance.value.id)
    }
    showWarning.value = false
  }

  function closeWarning() {
    showWarning.value = false
  }

  function handleEscKey(event: KeyboardEvent) {
    if (event.key === 'Escape' && showWarning.value) {
      showWarning.value = false
    }
  }

  return {
    showWarning,
    warningInstance,
    updateCountdown,
    extendFromWarning,
    closeWarning,
    handleEscKey,
  }
}
