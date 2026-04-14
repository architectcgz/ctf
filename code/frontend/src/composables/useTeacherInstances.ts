import { computed, reactive, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import { destroyTeacherInstance, getClasses, getTeacherInstances } from '@/api/teacher'
import type { TeacherClassItem, TeacherInstanceItem } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { useToast } from '@/composables/useToast'

type TeacherInstanceFilters = {
  className: string
  keyword: string
  studentNo: string
}

export function useTeacherInstances() {
  const pageSize = 20
  const authStore = useAuthStore()
  const toast = useToast()

  const classes = ref<TeacherClassItem[]>([])
  const instances = ref<TeacherInstanceItem[]>([])
  const page = ref(1)
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

  const isAdmin = computed(() => authStore.user?.role === 'admin')
  const totalCount = computed(() => instances.value.length)
  const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize)))
  const pagedInstances = computed(() => {
    const start = (page.value - 1) * pageSize
    return instances.value.slice(start, start + pageSize)
  })
  const runningCount = computed(
    () => instances.value.filter((item) => item.status === 'running').length
  )
  const expiringSoonCount = computed(
    () =>
      instances.value.filter((item) => item.status === 'running' && item.remaining_time <= 600)
        .length
  )

  async function initialize(): Promise<void> {
    loadingClasses.value = true
    error.value = null
    autoSearchReady.value = false
    page.value = 1

    try {
      classes.value = await getClasses()
      if (!isAdmin.value) {
        filters.className = authStore.user?.class_name || classes.value[0]?.name || ''
      }
      await loadInstances()
      autoSearchReady.value = true
    } catch (err) {
      console.error('加载教师实例管理页失败:', err)
      error.value = '加载实例管理数据失败，请稍后重试'
      classes.value = []
      instances.value = []
    } finally {
      loadingClasses.value = false
    }
  }

  async function loadInstances(): Promise<void> {
    const requestID = ++latestInstanceRequestID
    loadingInstances.value = true
    error.value = null

    try {
      const nextInstances = await getTeacherInstances({
        class_name: filters.className || undefined,
        keyword: filters.keyword.trim() || undefined,
        student_no: filters.studentNo.trim() || undefined,
      })
      if (requestID !== latestInstanceRequestID) return
      instances.value = nextInstances
      page.value = Math.min(page.value, Math.max(1, Math.ceil(nextInstances.length / pageSize)))
    } catch (err) {
      if (requestID !== latestInstanceRequestID) return
      console.error('加载教师实例列表失败:', err)
      error.value = '加载实例列表失败，请稍后重试'
      instances.value = []
    } finally {
      if (requestID !== latestInstanceRequestID) return
      loadingInstances.value = false
    }
  }

  type DebouncedInstanceSearch = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const scheduleInstanceSearch = useDebounceFn(() => {
    void loadInstances()
  }, 250) as DebouncedInstanceSearch

  async function submitFilters(): Promise<void> {
    scheduleInstanceSearch.cancel?.()
    page.value = 1
    await loadInstances()
  }

  async function resetFilters(): Promise<void> {
    autoSearchReady.value = false
    scheduleInstanceSearch.cancel?.()
    page.value = 1
    filters.keyword = ''
    filters.studentNo = ''
    filters.className = isAdmin.value
      ? ''
      : authStore.user?.class_name || classes.value[0]?.name || ''
    await loadInstances()
    autoSearchReady.value = true
  }

  function updateFilter<K extends keyof TeacherInstanceFilters>(
    key: K,
    value: TeacherInstanceFilters[K]
  ): void {
    filters[key] = value
  }

  function changePage(nextPage: number): void {
    page.value = Math.min(Math.max(1, nextPage), totalPages.value)
  }

  async function removeInstance(id: string): Promise<void> {
    destroyingId.value = id
    try {
      await destroyTeacherInstance(id)
      instances.value = instances.value.filter((item) => item.id !== id)
      page.value = Math.min(page.value, totalPages.value)
      toast.success('实例已销毁')
    } catch (err) {
      console.error('教师销毁实例失败:', err)
      const message = err instanceof Error && err.message.trim() ? err.message : '销毁实例失败，请稍后重试'
      toast.error(message)
    } finally {
      destroyingId.value = ''
    }
  }

  watch(
    () => [filters.className, filters.keyword, filters.studentNo],
    () => {
      if (!autoSearchReady.value) return
      page.value = 1
      scheduleInstanceSearch()
    }
  )

  return {
    classes,
    instances,
    pagedInstances,
    page,
    pageSize,
    totalPages,
    filters,
    loadingClasses,
    loadingInstances,
    destroyingId,
    error,
    isAdmin,
    totalCount,
    runningCount,
    expiringSoonCount,
    initialize,
    loadInstances,
    submitFilters,
    resetFilters,
    updateFilter,
    changePage,
    removeInstance,
  }
}
