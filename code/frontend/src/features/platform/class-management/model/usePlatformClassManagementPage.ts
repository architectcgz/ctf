import { computed } from 'vue'

import { platformClassStudentsRoute } from './platformClassManagementRoutes'
import { usePlatformClassDirectory } from './usePlatformClassDirectory'

export function usePlatformClassManagementPage() {
  const directory = usePlatformClassDirectory()

  function buildClassRoute(className: string) {
    return platformClassStudentsRoute(className)
  }

  const totalStudents = computed(() =>
    directory.list.value.reduce((sum, item) => sum + (item.student_count || 0), 0)
  )

  const rows = computed(() =>
    directory.list.value.map((item, index) => ({
      id: item.name,
      name: item.name,
      student_count: item.student_count || 0,
      teacher_name: '--',
      created_at: '--',
      actions: '查看班级',
      rowIndex: index,
    }))
  )

  return {
    list: directory.list,
    total: directory.total,
    page: directory.page,
    totalPages: directory.totalPages,
    loading: directory.loading,
    error: directory.error,
    totalStudents,
    rows,
    loadClasses: directory.loadClasses,
    handlePageChange: directory.handlePageChange,
    buildClassRoute,
  }
}
