import { computed, onMounted, onUnmounted, ref } from 'vue'

import {
  destroyInstance as apiDestroyInstance,
  extendInstance,
  getMyInstances,
  requestInstanceAccess,
} from '@/api/instance'
import type { InstanceListItem, InstanceStatus } from '@/api/contracts'
import { useClipboard } from '@/composables/useClipboard'
import { useToast } from '@/composables/useToast'

export const MAX_INSTANCES = 3
export const WARNING_THRESHOLD_SECONDS = 300
export const EXTEND_DURATION_SECONDS = 1800

export interface InstanceViewModel extends InstanceListItem {
  remaining: number
}

function isSharedInstance(instance: Pick<InstanceListItem, 'share_scope'>): boolean {
  return instance.share_scope === 'shared'
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
    pending: 'text-[#f59e0b]',
    creating: 'text-[#f59e0b]',
    running: 'text-[#22c55e]',
    expired: 'text-[var(--color-text-muted)]',
    destroying: 'text-[#f59e0b]',
    destroyed: 'text-[var(--color-text-muted)]',
    failed: 'text-[#ef4444]',
    crashed: 'text-[#ef4444]',
  }

  return classes[status] || 'text-[var(--color-text-muted)]'
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

export function useInstanceListPage() {
  const toast = useToast()
  const { copy } = useClipboard()

  const loading = ref(false)
  const instances = ref<InstanceViewModel[]>([])
  const showWarning = ref(false)
  const warningInstance = ref<InstanceViewModel | null>(null)
  const warnedInstances = new Set<string>()

  let timer: number | null = null

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

  async function loadInstances() {
    const data = await getMyInstances()
    instances.value = data.map(toViewModel)
  }

  async function refresh() {
    loading.value = true
    try {
      await loadInstances()
    } catch (error) {
      console.error('加载实例失败:', error)
      toast.error('加载实例失败，请刷新重试')
    } finally {
      loading.value = false
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
    if (target && isSharedInstance(target)) {
      toast.error('共享实例不支持手动延时')
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
      window.open(result.access_url, '_blank', 'noopener,noreferrer')
    } catch (error) {
      console.error('打开目标失败:', error)
      toast.error('打开目标失败，请稍后重试')
    }
  }

  async function destroyInstance(id: string) {
    const target = instances.value.find((instance) => instance.id === id)
    if (target && isSharedInstance(target)) {
      toast.error('共享实例不支持手动销毁')
      return
    }
    if (!window.confirm('确定要销毁该实例吗？此操作不可恢复。')) {
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
      const message = error instanceof Error && error.message.trim() ? error.message : '销毁失败，请稍后重试'
      toast.error(message)
    }
  }

  async function extendFromWarning() {
    if (warningInstance.value) {
      await extendTime(warningInstance.value.id)
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

  function updateCountdown() {
    const now = Date.now()

    instances.value = instances.value.map((instance) => {
      if (instance.status !== 'running') {
        return instance
      }
      if (isSharedInstance(instance)) {
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

      if (remaining < WARNING_THRESHOLD_SECONDS && !warnedInstances.has(instance.id)) {
        warnedInstances.add(instance.id)
        warningInstance.value = next
        showWarning.value = true
      }

      return next
    })
  }

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
