import { computed, onUnmounted, reactive, ref, watch } from 'vue'

import { destroyTeacherInstance, getClasses, getTeacherInstances } from '@/api/teacher'
import type { ClassDirectoryItem, TeacherInstanceItem } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { useAbortController } from '@/composables/useAbortController'
import { useToast } from '@/composables/useToast'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { reportFrontendError } from '@/utils/reportFrontendError'

type TeacherInstanceFilters = {
  className: string
  keyword: string
  studentNo: string
}

export function useInstances() {
  const authStore = useAuthStore()
  const toast = useToast()

  const classes = ref<ClassDirectoryItem[]>([])
  const instances = ref<TeacherInstanceItem[]>([])
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const filters = reactive<TeacherInstanceFilters>({
    className: '',
    keyword: '',
    studentNo: '',
  })

  const loadingClasses = ref(false)
  const loadingInstances = ref(false)
  const destroyingId = ref('')
  const error = ref<string | null>(null)
  const autoSearchReady = ref(false)
  let latestInstanceRequestID = 0
  let instanceSearchTimer: number | null = null
  const { createController, abort } = useAbortController()

  const isAdmin = computed(() => authStore.user?.role === 'admin')
  const totalCount = ref(0)
  const runningCount = ref(0)
  const expiringSoonCount = ref(0)
  const totalPages = computed(() =>
    Math.max(1, Math.ceil(totalCount.value / Math.max(pageSize.value, 1)))
  )

  async function initialize(): Promise<void> {
    loadingClasses.value = true
    error.value = null
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
      totalCount.value = 0
      runningCount.value = 0
      expiringSoonCount.value = 0
    } finally {
      loadingClasses.value = false
    }
  }

  async function loadInstances(): Promise<void> {
    const requestID = ++latestInstanceRequestID
    const controller = createController()
    loadingInstances.value = true
    error.value = null

    try {
      const nextInstances = await getTeacherInstances({
        class_name: filters.className || undefined,
        keyword: filters.keyword.trim() || undefined,
        student_no: filters.studentNo.trim() || undefined,
        page: page.value,
        page_size: pageSize.value,
      }, {
        signal: controller.signal,
      })
      if (requestID !== latestInstanceRequestID) return
      instances.value = nextInstances.list
      totalCount.value = nextInstances.total
      page.value = nextInstances.page
      pageSize.value = nextInstances.page_size
      runningCount.value = nextInstances.summary.running_count
      expiringSoonCount.value = nextInstances.summary.expiring_soon_count
    } catch (err) {
      if (requestID !== latestInstanceRequestID) return
      if (
        err &&
        typeof err === 'object' &&
        'code' in err &&
        (err as { code?: unknown }).code === 'ERR_CANCELED'
      ) {
        return
      }
      reportFrontendError('加载教师实例列表失败:', err)
      error.value = '加载实例列表失败，请稍后重试'
      instances.value = []
      totalCount.value = 0
      runningCount.value = 0
      expiringSoonCount.value = 0
    } finally {
      if (requestID !== latestInstanceRequestID) return
      loadingInstances.value = false
    }
  }

  function clearScheduledInstanceSearch(): void {
    if (instanceSearchTimer !== null) {
      window.clearTimeout(instanceSearchTimer)
      instanceSearchTimer = null
    }
  }

  function scheduleInstanceSearch(): void {
    clearScheduledInstanceSearch()
    instanceSearchTimer = window.setTimeout(() => {
      instanceSearchTimer = null
      page.value = 1
      void loadInstances()
    }, 250)
  }

  function updateFilter<K extends keyof TeacherInstanceFilters>(
    key: K,
    value: TeacherInstanceFilters[K]
  ): void {
    filters[key] = value
  }

  async function removeInstance(id: string): Promise<void> {
    destroyingId.value = id
    try {
      await destroyTeacherInstance(id)
      if (instances.value.length === 1 && page.value > 1) {
        page.value -= 1
      }
      await loadInstances()
      toast.success('实例已销毁')
    } catch (err) {
      reportFrontendError('教师销毁实例失败:', err)
      const message =
        err instanceof Error && err.message.trim() ? err.message : '销毁实例失败，请稍后重试'
      toast.error(message)
    } finally {
      destroyingId.value = ''
    }
  }

  watch(
    () => [filters.className, filters.keyword, filters.studentNo],
    () => {
      if (!autoSearchReady.value) return
      scheduleInstanceSearch()
    }
  )

  onUnmounted(() => {
    clearScheduledInstanceSearch()
    abort()
  })

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
