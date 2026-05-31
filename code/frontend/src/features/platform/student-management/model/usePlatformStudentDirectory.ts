import { computed, onMounted, ref } from 'vue'

import { getClasses, getStudentsDirectory } from '@/api/admin'
import type { ClassDirectoryItem } from '@/api/contracts'
import { useStudentDirectoryQuery } from '@/features/student-directory'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { reportFrontendError } from '@/utils/reportFrontendError'

export function usePlatformStudentDirectory() {
  const classes = ref<ClassDirectoryItem[]>([])
  const loadingClasses = ref(false)
  const pageError = ref<string | null>(null)
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const keyword = ref('')
  const classFilter = ref('')
  const studentDirectoryQuery = useStudentDirectoryQuery({
    debounceMs: 250,
    errorMessage: '加载学生目录失败，请稍后重试',
    request: getStudentsDirectory,
  })

  const list = computed(() => studentDirectoryQuery.students.value)
  const total = computed(() => studentDirectoryQuery.total.value)
  const loading = computed(() => studentDirectoryQuery.loading.value)
  const error = computed(() => pageError.value ?? studentDirectoryQuery.error.value)
  const hasActiveFilters = computed(() => Boolean(keyword.value.trim() || classFilter.value))
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))
  const activeStudents = computed(() =>
    list.value.filter((item) => (item.recent_event_count ?? 0) > 0).length
  )
  const assignedClassCount = computed(() =>
    classes.value.filter((item) => (item.student_count ?? 0) > 0).length
  )
  const directoryParams = computed(() => ({
    class_name: classFilter.value || undefined,
    keyword: keyword.value.trim() || undefined,
    student_no: undefined,
    sort_key: 'name' as const,
    sort_order: 'asc' as const,
    page: page.value,
    page_size: pageSize.value,
  }))

  async function loadClasses(): Promise<void> {
    loadingClasses.value = true
    try {
      classes.value = await getClasses()
    } finally {
      loadingClasses.value = false
    }
  }

  async function loadStudents(): Promise<void> {
    await studentDirectoryQuery.loadStudents(directoryParams.value)
  }

  async function initialize(): Promise<void> {
    pageError.value = null
    studentDirectoryQuery.cancelScheduledLoad()

    try {
      await loadClasses()
      await loadStudents()
    } catch (err) {
      reportFrontendError('初始化学生管理失败:', err)
      pageError.value = '加载学生管理失败，请稍后重试'
    }
  }

  function handleKeywordChange(value: string): void {
    keyword.value = value
    page.value = 1
    studentDirectoryQuery.scheduleLoadStudents({
      ...directoryParams.value,
      keyword: value.trim() || undefined,
      page: 1,
    })
  }

  function handleClassFilterChange(value: string): void {
    classFilter.value = value
    page.value = 1
    studentDirectoryQuery.cancelScheduledLoad()
    void studentDirectoryQuery.loadStudents({
      ...directoryParams.value,
      class_name: value || undefined,
      page: 1,
    })
  }

  function resetFilters(): void {
    keyword.value = ''
    classFilter.value = ''
    page.value = 1
    studentDirectoryQuery.cancelScheduledLoad()
    void studentDirectoryQuery.loadStudents({
      ...directoryParams.value,
      class_name: undefined,
      keyword: undefined,
      page: 1,
    })
  }

  function handlePageChange(nextPage: number): void {
    const normalizedPage = Math.max(1, Math.floor(nextPage))
    if (normalizedPage === page.value || normalizedPage > totalPages.value) {
      return
    }

    page.value = normalizedPage
    void loadStudents()
  }

  onMounted(() => {
    void initialize()
  })

  return {
    classes,
    loadingClasses,
    keyword,
    classFilter,
    list,
    total,
    loading,
    error,
    hasActiveFilters,
    page,
    totalPages,
    activeStudents,
    assignedClassCount,
    initialize,
    handleKeywordChange,
    handleClassFilterChange,
    resetFilters,
    handlePageChange,
  }
}
