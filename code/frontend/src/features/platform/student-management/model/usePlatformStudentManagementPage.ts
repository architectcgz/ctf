import { computed } from 'vue'

import { platformStudentAnalysisRoute } from './platformStudentManagementRoutes'
import { usePlatformStudentDirectory } from './usePlatformStudentDirectory'

export function usePlatformStudentManagementPage() {
  const directory = usePlatformStudentDirectory()
  const rows = computed(() =>
    directory.list.value.map((item) => ({
      id: item.id,
      name: item.name?.trim() || '未设置姓名',
      username: item.username,
      student_no: item.student_no?.trim() || '未设置学号',
      class_name: item.class_name || '未分班',
      total_score: item.total_score ?? 0,
      actions: '查看学员',
      studentRoute: buildStudentRoute(item.id, item.class_name),
    }))
  )

  function buildStudentRoute(studentId: string, className?: string) {
    return platformStudentAnalysisRoute(studentId, className || directory.classFilter.value || '')
  }

  return {
    classes: directory.classes,
    loadingClasses: directory.loadingClasses,
    keyword: directory.keyword,
    classFilter: directory.classFilter,
    list: directory.list,
    total: directory.total,
    loading: directory.loading,
    error: directory.error,
    hasActiveFilters: directory.hasActiveFilters,
    page: directory.page,
    totalPages: directory.totalPages,
    activeStudents: directory.activeStudents,
    assignedClassCount: directory.assignedClassCount,
    rows,
    initialize: directory.initialize,
    handleKeywordChange: directory.handleKeywordChange,
    handleClassFilterChange: directory.handleClassFilterChange,
    resetFilters: directory.resetFilters,
    handlePageChange: directory.handlePageChange,
    buildStudentRoute,
  }
}
