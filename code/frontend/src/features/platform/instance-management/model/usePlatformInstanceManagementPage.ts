import { computed, onMounted, ref, watch } from 'vue'

import type { InstanceDirectoryItem } from '@/api/contracts'
import { getInstanceStudentDisplayName } from '@/entities/instance'
import { useManagedInstanceDirectory } from '@/features/managed-instance-directory'
import { useManagedInstanceDestroyAction } from '@/features/managed-instance-workflow'
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
  class_name: string
  ip_address: string
  status: string
  created_at: string
  actions: string
  studentRoute: {
    name: string
    params?: Record<string, string>
  }
}

type InstanceStatusFilter = 'running' | 'creating' | 'expired' | 'failed' | 'inactive' | ''

export function usePlatformInstanceManagementPage() {
  const keyword = ref('')
  const statusFilter = ref<InstanceStatusFilter>('')
  const runningCount = ref(0)
  const warningCount = ref(0)
  const {
    list,
    total,
    page,
    pageSize,
    loading,
    error,
    totalPages,
    loadInstances,
    scheduleSearch,
    handlePageChange,
  } = useManagedInstanceDirectory({
    role: 'admin',
    initialPageSize: 15,
    buildQuery: ({ page, pageSize }) => ({
      class_name: undefined,
      keyword: keyword.value.trim() || undefined,
      student_no: undefined,
      status: statusFilter.value || undefined,
      page,
      page_size: pageSize,
    }),
    errorMessage: '加载实例列表失败，请稍后重试',
    onLoaded: (response) => {
      runningCount.value = response.summary.running_count
      warningCount.value = response.summary.warning_count
    },
    onLoadError: (err) => {
      reportFrontendError('加载实例列表失败:', err)
      runningCount.value = 0
      warningCount.value = 0
    },
  })

  const totalInstances = computed(() => total.value)
  const filteredTotal = computed(() => total.value)
  const overviewRoute = platformOverviewRoute
  const pageRows = computed<InstanceManageTableRow[]>(() => {
    return list.value.map((item) => ({
      id: item.id,
      challenge: item.challenge_title,
      student_id: String(item.student_id),
      user: getInstanceStudentDisplayName(item),
      class_name: item.class_name,
      ip_address: item.access_url || '暂未分配',
      status: item.status,
      created_at: formatDateTime(item.created_at),
      actions: '销毁',
      studentRoute: buildStudentRoute(String(item.student_id), item.class_name),
    }))
  })

  const { destroyingId, destroyManagedInstance } = useManagedInstanceDestroyAction({
    role: 'admin',
    resolveTarget: (id) =>
      list.value.find((instance) => instance.id === id) ?? ({ id } as InstanceDirectoryItem),
    buildConfirmOptions: (instance) => ({
      title: '强制销毁实例',
      message: `您确定要强制销毁实例 ${instance.id} 吗？此操作不可逆，用户当前的运行状态将丢失。`,
      confirmButtonText: '强制销毁',
      cancelButtonText: '取消',
    }),
    onDestroyed: async () => {
      if (list.value.length === 1 && page.value > 1) {
        page.value -= 1
      }
      await loadInstances()
    },
    onDestroyError: ({ error: err }) => {
      reportFrontendError('销毁实例失败:', err)
      error.value = '销毁实例失败，请稍后重试'
    },
  })

  function requestDestroyById(id: string): void {
    void destroyManagedInstance(id)
  }

  function buildStudentRoute(studentId: string, className: string) {
    return platformInstanceStudentAnalysisRoute(studentId, className)
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

  watch([keyword, statusFilter], () => {
    scheduleSearch()
  })

  onMounted(() => {
    void loadInstances()
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
    handlePageChange: (pageNumber: number) => {
      void handlePageChange(pageNumber)
    },
    setKeyword,
    setStatusFilter,
    resetFilters,
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
