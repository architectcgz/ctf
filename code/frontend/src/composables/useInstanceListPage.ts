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
    failed: '失败',
    crashed: '崩溃',
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
      toast.error('销毁失败，请稍后重试')
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
