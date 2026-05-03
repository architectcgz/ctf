import { computed, onMounted, onUnmounted, ref } from 'vue'

import { getMyInstances } from '@/api/instance'
import type { InstanceListItem, InstanceStatus } from '@/api/contracts'
import { useToast } from '@/composables/useToast'
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

export function getInstanceStatusLabel(status: InstanceStatus): string {
  const labels: Record<InstanceStatus, string> = {
    pending: '等待中',
    creating: '创建中',
    running: '运行中',
    expired: '已过期',
    destroying: '销毁中',
    destroyed: '已销毁',
    failed: '启动失败',
    crashed: '运行异常',
  }

  return labels[status] || status
}

export function getInstanceStatusClass(status: InstanceStatus): string {
  const classes: Record<InstanceStatus, string> = {
    pending: 'instance-status-dot--warning',
    creating: 'instance-status-dot--warning',
    running: 'instance-status-dot--success',
    expired: 'instance-status-dot--muted',
    destroying: 'instance-status-dot--warning',
    destroyed: 'instance-status-dot--muted',
    failed: 'instance-status-dot--danger',
    crashed: 'instance-status-dot--danger',
  }

  return classes[status] || 'instance-status-dot--muted'
}

export function formatRemainingTime(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
}

function formatEtaSeconds(seconds?: number): string {
  if (typeof seconds !== 'number' || seconds <= 0) return '预计时间计算中'
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  if (minutes <= 0) return `${secs} 秒`
  return `${minutes} 分 ${secs} 秒`
}

export function getInstanceWaitingHint(
  instance: Pick<InstanceListItem, 'status' | 'queue_position' | 'eta_seconds' | 'progress'>
): string {
  if (instance.status === 'failed') {
    return '启动失败，当前目标不可访问'
  }
  if (instance.status === 'crashed') {
    return '实例运行异常，当前目标不可访问'
  }
  if (instance.status !== 'pending' && instance.status !== 'creating') {
    return ''
  }

  const details: string[] = ['实例正在排队创建']

  if (typeof instance.queue_position === 'number' && instance.queue_position > 0) {
    details.push(`队列第 ${instance.queue_position} 位`)
  }

  details.push(`预计等待 ${formatEtaSeconds(instance.eta_seconds)}`)

  if (typeof instance.progress === 'number') {
    const progress = Math.max(0, Math.min(100, Math.round(instance.progress)))
    details.push(`进度 ${progress}%`)
  }

  return details.join('，')
}

export function formatInstanceAccessDisplay(
  instance: Pick<InstanceListItem, 'access_url' | 'access' | 'ssh_info'>
): string {
  return (
    instance.access?.command ||
    instance.access_url ||
    (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : '')
  )
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
