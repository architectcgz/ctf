import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

import type { InstanceDirectoryItem } from '@/api/contracts'
import {
  destroyManagedInstanceByRole,
  getInstanceDirectoryByRole,
} from '@/api/instances'
import { useAbortController } from '@/shared/lib/request/useAbortController'
import { confirmDestructiveAction } from '@/shared/model/common/useDestructiveConfirm'
import { reportFrontendError } from '@/utils/reportFrontendError'
import {
  platformInstanceStudentAnalysisRoute,
  platformOverviewRoute,
} from './platformInstanceManagementRoutes'

interface InstanceManageTableRow {
  id: string
  challenge: string
  student_id: string
  user: string
  username: string
  class_name: string
  ip_address: string
  status: string
  status_label: string
  created_at: string
  actions: string
  studentRoute: {
    name: string
    params?: Record<string, string>
  }
}

type InstanceStatusFilter = 'running' | 'creating' | 'expired' | 'failed' | 'inactive' | ''

export function usePlatformInstanceManagementPage() {
  const list = ref<InstanceDirectoryItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(15)
  const loading = ref(false)
  const destroyingId = ref('')
  const error = ref<string | null>(null)
  const keyword = ref('')
  const statusFilter = ref<InstanceStatusFilter>('')
  let latestRequestId = 0
  let scheduledSearchTimer: number | null = null
  const { createController, abort } = useAbortController()

  const totalInstances = computed(() => total.value)
  const filteredTotal = computed(() => total.value)
  const overviewRoute = platformOverviewRoute
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))
  const pageRows = computed<InstanceManageTableRow[]>(() => {
    return list.value.map((item) => ({
      id: item.id,
      challenge: item.challenge_title,
      student_id: String(item.student_id),
      user: item.student_name || item.student_username,
      username: item.student_username,
      class_name: item.class_name,
      ip_address: item.access_url || '暂未分配',
      status: item.status,
      status_label: formatStatus(item.status),
      created_at: formatDateTime(item.created_at),
      actions: '销毁',
      studentRoute: buildStudentRoute(String(item.student_id), item.class_name),
    }))
  })
  const runningCount = ref(0)
  const warningCount = ref(0)

  async function loadInstances(): Promise<void> {
    const requestId = ++latestRequestId
    const controller = createController()
    loading.value = true
    error.value = null
    try {
      const response = await getInstanceDirectoryByRole(
        'admin',
        {
          class_name: undefined,
          keyword: keyword.value.trim() || undefined,
          student_no: undefined,
          status: statusFilter.value || undefined,
          page: page.value,
          page_size: pageSize.value,
        },
        {
          signal: controller.signal,
        }
      )
      if (requestId !== latestRequestId) return
      list.value = response.list
      total.value = response.total
      page.value = response.page
      pageSize.value = response.page_size
      runningCount.value = response.summary.running_count
      warningCount.value = response.summary.warning_count
      if (page.value > totalPages.value) {
        page.value = totalPages.value
      }
    } catch (err) {
      if (requestId !== latestRequestId) return
      if (
        err &&
        typeof err === 'object' &&
        'code' in err &&
        (err as { code?: unknown }).code === 'ERR_CANCELED'
      ) {
        return
      }
      reportFrontendError('加载实例列表失败:', err)
      error.value = '加载实例列表失败，请稍后重试'
      list.value = []
      total.value = 0
      runningCount.value = 0
      warningCount.value = 0
    } finally {
      if (requestId !== latestRequestId) return
      loading.value = false
    }
  }

  async function handleDestroyInstance(instance: InstanceDirectoryItem): Promise<void> {
    const confirmed = await confirmDestructiveAction({
      title: '强制销毁实例',
      message: `您确定要强制销毁实例 ${instance.id} 吗？此操作不可逆，用户当前的运行状态将丢失。`,
      confirmButtonText: '强制销毁',
      cancelButtonText: '取消',
    })

    if (!confirmed) return

    try {
      destroyingId.value = instance.id
      await destroyManagedInstanceByRole('admin', instance.id)
      if (list.value.length === 1 && page.value > 1) {
        page.value -= 1
      }
      await loadInstances()
    } catch (err) {
      reportFrontendError('销毁实例失败:', err)
      error.value = '销毁实例失败，请稍后重试'
    } finally {
      destroyingId.value = ''
    }
  }

  function requestDestroyById(id: string): void {
    const instance = list.value.find((item) => item.id === id)
    if (!instance) {
      return
    }

    void handleDestroyInstance(instance)
  }

  function buildStudentRoute(studentId: string, className: string) {
    return platformInstanceStudentAnalysisRoute(studentId, className)
  }

  function handlePageChange(p: number): void {
    const normalizedPage = Math.max(1, Math.floor(p))
    if (normalizedPage === page.value || normalizedPage > totalPages.value) {
      return
    }

    page.value = normalizedPage
    void loadInstances()
  }

  function setKeyword(nextKeyword: string): void {
    keyword.value = nextKeyword
  }

  function setStatusFilter(nextStatusFilter: InstanceStatusFilter): void {
    statusFilter.value = nextStatusFilter
  }

  function resetFilters(): void {
    keyword.value = ''
    statusFilter.value = ''
  }

  function clearScheduledSearch(): void {
    if (scheduledSearchTimer !== null) {
      window.clearTimeout(scheduledSearchTimer)
      scheduledSearchTimer = null
    }
  }

  function scheduleSearch(): void {
    clearScheduledSearch()
    scheduledSearchTimer = window.setTimeout(() => {
      scheduledSearchTimer = null
      page.value = 1
      void loadInstances()
    }, 250)
  }

  watch([keyword, statusFilter], () => {
    scheduleSearch()
  })

  onMounted(() => {
    void loadInstances()
  })

  onUnmounted(() => {
    clearScheduledSearch()
    abort()
  })

  return {
    list,
    page,
    loading,
    destroyingId,
    error,
    keyword,
    statusFilter,
    totalInstances,
    filteredTotal,
    overviewRoute,
    totalPages,
    pageRows,
    runningCount,
    warningCount,
    loadInstances,
    buildStudentRoute,
    requestDestroyById,
    handlePageChange,
    setKeyword,
    setStatusFilter,
    resetFilters,
  }
}

function formatStatus(status: string): string {
  switch (status) {
    case 'running':
      return '运行中'
    case 'creating':
      return '创建中'
    case 'expired':
      return '已过期'
    case 'failed':
      return '异常'
    default:
      return status
  }
}

function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '--'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
