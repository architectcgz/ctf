import { computed, reactive, ref, watch } from 'vue'

import { getClasses } from '@/api/teacher'
import type { ClassDirectoryItem, InstanceDirectoryItem } from '@/api/contracts'
import { useManagedInstanceDirectory } from '@/features/managed-instance-directory'
import { useManagedInstanceDestroyAction } from '@/features/managed-instance-workflow'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/shared/model/common/useToast'
import { reportFrontendError } from '@/utils/reportFrontendError'

type TeacherInstanceFilters = {
  className: string
  keyword: string
  studentNo: string
}

export function useTeacherInstanceDirectoryState() {
  const authStore = useAuthStore()
  const toast = useToast()

  const classes = ref<ClassDirectoryItem[]>([])
  const filters = reactive<TeacherInstanceFilters>({
    className: '',
    keyword: '',
    studentNo: '',
  })

  const loadingClasses = ref(false)
  const autoSearchReady = ref(false)
  const totalCount = ref(0)
  const runningCount = ref(0)
  const expiringSoonCount = ref(0)
  const {
    list: instances,
    page,
    pageSize,
    loading: loadingInstances,
    error,
    totalPages,
    total,
    loadInstances,
    scheduleSearch: scheduleInstanceSearch,
  } = useManagedInstanceDirectory({
    role: () => authStore.user?.role,
    buildQuery: ({ page, pageSize }) => ({
      class_name: filters.className || undefined,
      keyword: filters.keyword.trim() || undefined,
      student_no: filters.studentNo.trim() || undefined,
      page,
      page_size: pageSize,
    }),
    errorMessage: '加载实例列表失败，请稍后重试',
    onLoaded: (response) => {
      totalCount.value = response.total
      runningCount.value = response.summary.running_count
      expiringSoonCount.value = response.summary.expiring_soon_count
    },
    onLoadError: (err) => {
      reportFrontendError('加载教师实例列表失败:', err)
      totalCount.value = 0
      runningCount.value = 0
      expiringSoonCount.value = 0
    },
  })

  const isAdmin = computed(() => authStore.user?.role === 'admin')

  async function initialize(): Promise<void> {
    loadingClasses.value = true
    autoSearchReady.value = false

    try {
      classes.value = await getClasses()
      if (!isAdmin.value) {
        filters.className = authStore.user?.class_name || classes.value[0]?.name || ''
      }
      page.value = 1
      await loadInstances()
      autoSearchReady.value = true
    } catch (err) {
      reportFrontendError('加载教师实例管理页失败:', err)
      error.value = '加载实例管理数据失败，请稍后重试'
      classes.value = []
      instances.value = []
      total.value = 0
      totalCount.value = 0
      runningCount.value = 0
      expiringSoonCount.value = 0
    } finally {
      loadingClasses.value = false
    }
  }

  function updateFilter<K extends keyof TeacherInstanceFilters>(
    key: K,
    value: TeacherInstanceFilters[K]
  ): void {
    filters[key] = value
  }

  const { destroyingId, destroyManagedInstance: removeInstance } = useManagedInstanceDestroyAction({
    role: () => authStore.user?.role,
    resolveTarget: (id) =>
      instances.value.find((instance) => instance.id === id) ??
      ({ id } as InstanceDirectoryItem),
    buildConfirmOptions: () => ({
      title: '确认销毁实例',
      message: '确定要销毁该实例吗？此操作不可恢复。',
      confirmButtonText: '确认销毁',
      cancelButtonText: '取消',
    }),
    onDestroyed: async () => {
      if (instances.value.length === 1 && page.value > 1) {
        page.value -= 1
      }
      await loadInstances()
      toast.success('实例已销毁')
    },
    onDestroyError: ({ error: err, message }) => {
      reportFrontendError('教师销毁实例失败:', err)
      toast.error(message)
    },
    fallbackErrorMessage: '销毁实例失败，请稍后重试',
  })

  watch(
    () => [filters.className, filters.keyword, filters.studentNo],
    () => {
      if (!autoSearchReady.value) return
      scheduleInstanceSearch()
    }
  )

  return {
    classes,
    instances,
    page,
    pageSize,
    filters,
    loadingClasses,
    loadingInstances,
    destroyingId,
    error,
    isAdmin,
    totalCount,
    runningCount,
    expiringSoonCount,
    totalPages,
    initialize,
    loadInstances,
    updateFilter,
    removeInstance,
  }
}
