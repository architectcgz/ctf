import { computed, onMounted, onUnmounted, ref } from 'vue'

import {
  destroyInstance as apiDestroyInstance,
  extendInstance,
  getMyInstances,
  requestInstanceAccess,
} from '@/api/instance'
import type { InstanceListItem, InstanceStatus } from '@/api/contracts'
import { useClipboard } from '@/composables/useClipboard'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'
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
  const { copy } = useClipboard()

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
    onExtendInstance: extendTime,
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
