import { ref } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import { getStudentsDirectory, type TeacherStudentDirectoryParams } from '@/api/teacher'
import type { PageResult, TeacherStudentItem } from '@/api/contracts'

interface UseStudentDirectoryQueryOptions {
  debounceMs?: number
  errorMessage: string
  request?: (
    params?: TeacherStudentDirectoryParams
  ) => Promise<PageResult<TeacherStudentItem>>
}

export function useStudentDirectoryQuery(options: UseStudentDirectoryQueryOptions) {
  const students = ref<TeacherStudentItem[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  let latestRequestId = 0
  const requestStudents = options.request ?? getStudentsDirectory
  type DebouncedStudentLoader = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const debouncedLoadStudents = options.debounceMs
    ? (useDebounceFn((params?: TeacherStudentDirectoryParams) => {
        void loadStudents(params)
      }, options.debounceMs) as DebouncedStudentLoader)
    : null

  async function loadStudents(
    params?: TeacherStudentDirectoryParams
  ): Promise<PageResult<TeacherStudentItem>> {
    const requestId = ++latestRequestId
    loading.value = true
    error.value = null

    try {
      const nextPage = await requestStudents(params)
      if (requestId !== latestRequestId) {
        return {
          list: students.value,
          total: total.value,
          page: params?.page ?? 1,
          page_size: params?.page_size ?? nextPage.page_size,
        }
      }
      students.value = nextPage.list
      total.value = nextPage.total
      return nextPage
    } catch (err) {
      if (requestId !== latestRequestId) {
        return {
          list: students.value,
          total: total.value,
          page: params?.page ?? 1,
          page_size: params?.page_size ?? 20,
        }
      }
      console.error('加载学生目录失败:', err)
      error.value = options.errorMessage
      students.value = []
      total.value = 0
      return {
        list: [],
        total: 0,
        page: params?.page ?? 1,
        page_size: params?.page_size ?? 20,
      }
    } finally {
      if (requestId === latestRequestId) {
        loading.value = false
      }
    }
  }

  function scheduleLoadStudents(params?: TeacherStudentDirectoryParams): void {
    if (debouncedLoadStudents) {
      debouncedLoadStudents(params)
      return
    }

    void loadStudents(params)
  }

  function cancelScheduledLoad(): void {
    debouncedLoadStudents?.cancel?.()
  }

  function clearStudents(): void {
    latestRequestId += 1
    students.value = []
    total.value = 0
    error.value = null
    loading.value = false
  }

  return {
    students,
    total,
    loading,
    error,
    loadStudents,
    scheduleLoadStudents,
    cancelScheduledLoad,
    clearStudents,
  }
}
