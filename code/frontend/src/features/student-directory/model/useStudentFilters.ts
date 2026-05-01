import { computed, ref } from 'vue'

export interface StudentQueryParams {
  keyword?: string
  student_no?: string
}

interface UseStudentFiltersOptions {
  selectedClassName?: string
  searchQuery?: string
  studentNoQuery?: string
}

export function useStudentFilters(options: UseStudentFiltersOptions = {}) {
  const selectedClassName = ref(options.selectedClassName ?? '')
  const searchQuery = ref(options.searchQuery ?? '')
  const studentNoQuery = ref(options.studentNoQuery ?? '')

  const studentQueryParams = computed<StudentQueryParams>(() => ({
    keyword: searchQuery.value.trim() || undefined,
    student_no: studentNoQuery.value.trim() || undefined,
  }))

  function updateSelectedClassName(value: string): void {
    selectedClassName.value = value
  }

  function updateSearchQuery(value: string): void {
    searchQuery.value = value
  }

  function updateStudentNoQuery(value: string): void {
    studentNoQuery.value = value
  }

  function resetTextFilters(): void {
    searchQuery.value = ''
    studentNoQuery.value = ''
  }

  return {
    selectedClassName,
    searchQuery,
    studentNoQuery,
    studentQueryParams,
    updateSelectedClassName,
    updateSearchQuery,
    updateStudentNoQuery,
    resetTextFilters,
  }
}
