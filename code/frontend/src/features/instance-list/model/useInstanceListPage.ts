import { computed, onMounted, onUnmounted, ref } from 'vue'

import { getMyInstances } from '@/api/instance'
import type { InstanceListItem } from '@/api/contracts'
import { useToast } from '@/shared/model/common/useToast'
import { useInstanceOperations } from './useInstanceOperations'
import { useInstanceWarningState } from './useInstanceWarningState'

export const MAX_INSTANCES = 3
export const WARNING_THRESHOLD_SECONDS = 300
export const EXTEND_DURATION_SECONDS = 1800
export const INSTANCE_STATUS_REFRESH_INTERVAL_MS = 5000

export interface InstanceViewModel extends InstanceListItem {
  remaining: number
}

function isSharedInstance(instance: Pick<InstanceListItem, 'share_scope'>): boolean {
  return instance.share_scope === 'shared'
}

function isAWDTeamInstance(
  instance: Pick<InstanceListItem, 'contest_mode' | 'share_scope'>
): boolean {
  return instance.contest_mode === 'awd' && instance.share_scope === 'per_team'
}

export function isInstanceManualActionAllowed(
  instance: Pick<InstanceListItem, 'contest_mode' | 'share_scope'>
): boolean {
  return !isSharedInstance(instance) && !isAWDTeamInstance(instance)
}

function calculateRemaining(expiresAt: string): number {
  return Math.max(0, Math.floor((new Date(expiresAt).getTime() - Date.now()) / 1000))
}

function toViewModel(item: InstanceListItem): InstanceViewModel {
  return {
    ...item,
    remaining: calculateRemaining(item.expires_at),
  }
}

export function canOpenInstanceInBrowser(
  instance: Pick<InstanceListItem, 'access_url' | 'access'>
): boolean {
  return Boolean(instance.access_url) && instance.access?.protocol !== 'tcp'
}

export function useInstanceListPage() {
  const toast = useToast()

  const loading = ref(false)
  const instances = ref<InstanceViewModel[]>([])
  const warnedInstances = new Set<string>()

  let timer: number | null = null
  let statusRefreshTimer: number | null = null
  let refreshInFlight = false

  const maxInstances = MAX_INSTANCES
  const runningCount = computed(
    () => instances.value.filter((instance) => instance.status === 'running').length
  )
  const waitingCount = computed(
    () =>
      instances.value.filter(
        (instance) => instance.status === 'pending' || instance.status === 'creating'
      ).length
  )

  function hasPendingRemoteStatus(instance: InstanceViewModel): boolean {
    return instance.status === 'pending' || instance.status === 'creating'
  }

  function stopStatusRefresh() {
    if (statusRefreshTimer !== null) {
      window.clearInterval(statusRefreshTimer)
      statusRefreshTimer = null
    }
  }

  function syncStatusRefresh() {
    const shouldPoll = instances.value.some(hasPendingRemoteStatus)
    if (!shouldPoll) {
      stopStatusRefresh()
      return
    }
    if (statusRefreshTimer !== null) {
      return
    }
    statusRefreshTimer = window.setInterval(() => {
      void refresh({ silent: true })
    }, INSTANCE_STATUS_REFRESH_INTERVAL_MS)
  }

  async function loadInstances() {
    const data = await getMyInstances()
    instances.value = data.map(toViewModel)
  }

  async function refresh(options?: { silent?: boolean }) {
    if (refreshInFlight) {
      return
    }

    refreshInFlight = true
    if (!options?.silent) {
      loading.value = true
    }
    try {
      await loadInstances()
    } catch (error) {
      if (!options?.silent) {
        console.error('加载实例失败:', error)
        toast.error('加载实例失败，请刷新重试')
      }
    } finally {
      refreshInFlight = false
      syncStatusRefresh()
      if (!options?.silent) {
        loading.value = false
      }
    }
  }

  const {
    showWarning,
    warningInstance,
    updateCountdown,
    extendFromWarning,
    closeWarning,
    handleEscKey,
  } = useInstanceWarningState({
    instances,
    warnedInstances,
    warningThresholdSeconds: WARNING_THRESHOLD_SECONDS,
    canManualAction: isInstanceManualActionAllowed,
    onExtendInstance: async (id) => {
      await extendTime(id)
    },
  })
  const { copyAddress, extendTime, openTarget, destroyInstance } = useInstanceOperations({
    instances,
    warnedInstances,
    warningInstance,
    showWarning,
    isInstanceManualActionAllowed,
    isAWDTeamInstance,
    calculateRemaining,
    loadInstances,
  })

  onMounted(() => {
    void refresh()
    timer = window.setInterval(updateCountdown, 1000)
    window.addEventListener('keydown', handleEscKey)
  })

  onUnmounted(() => {
    if (timer !== null) {
      window.clearInterval(timer)
      timer = null
    }
    stopStatusRefresh()
    window.removeEventListener('keydown', handleEscKey)
  })

  return {
    loading,
    maxInstances,
    instances,
    runningCount,
    waitingCount,
    showWarning,
    warningInstance,
    copyAddress,
    extendTime,
    openTarget,
    destroyInstance,
    extendFromWarning,
    closeWarning,
  }
}
