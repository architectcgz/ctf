import { ref } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import { getClassStudents } from '@/api/teacher'
import type { TeacherStudentItem } from '@/api/contracts'

import type { StudentQueryParams } from './useStudentFilters'

interface UseStudentListQueryOptions {
  debounceMs?: number
  errorMessage: string
  getParams?: () => StudentQueryParams | undefined
  request?: (className: string, params?: StudentQueryParams) => Promise<TeacherStudentItem[]>
}

export function useStudentListQuery(options: UseStudentListQueryOptions) {
  const students = ref<TeacherStudentItem[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  let latestRequestId = 0
  const requestStudents = options.request ?? getClassStudents
  type DebouncedStudentLoader = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const debouncedLoadStudents = options.debounceMs
    ? (useDebounceFn((className: string) => {
        void loadStudents(className)
      }, options.debounceMs) as DebouncedStudentLoader)
    : null

  async function loadStudents(className: string): Promise<TeacherStudentItem[]> {
    if (!className) {
      clearStudents()
      return []
    }

    const requestId = ++latestRequestId
    loading.value = true
    error.value = null

    try {
      const nextStudents = await requestStudents(className, options.getParams?.())
      if (requestId !== latestRequestId) {
        return students.value
      }
      students.value = nextStudents
      return nextStudents
    } catch (err) {
      if (requestId !== latestRequestId) {
        return students.value
      }
      console.error('加载学生列表失败:', err)
      error.value = options.errorMessage
      students.value = []
      return []
    } finally {
      if (requestId === latestRequestId) {
        loading.value = false
      }
    }
  }

  function scheduleLoadStudents(className: string): void {
    if (debouncedLoadStudents) {
      debouncedLoadStudents(className)
      return
    }

    void loadStudents(className)
  }

  function cancelScheduledLoad(): void {
    debouncedLoadStudents?.cancel?.()
  }

  function clearStudents(): void {
    latestRequestId += 1
    students.value = []
    error.value = null
    loading.value = false
  }

  return {
    students,
    loading,
    error,
    loadStudents,
    scheduleLoadStudents,
    cancelScheduledLoad,
    clearStudents,
  }
}
