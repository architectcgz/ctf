import { computed, ref } from 'vue'

import { useAuthStore } from '@/stores/auth'

import { teacherClassStudentsRoute, teacherDashboardRoute } from './teacherClassManagementRoutes'
import { useTeacherClassDirectory } from './useTeacherClassDirectory'

export function useClassManagementPage() {
  const authStore = useAuthStore()
  const reportDialogVisible = ref(false)
  const directory = useTeacherClassDirectory()
  const defaultReportClassName = computed(() => authStore.user?.class_name ?? '')

  function buildClassRoute(className: string) {
    return teacherClassStudentsRoute(className)
  }

  function openClassReportDialog(): void {
    reportDialogVisible.value = true
  }

  return {
    classes: directory.classes,
    total: directory.total,
    page: directory.page,
    pageSize: directory.pageSize,
    loading: directory.loading,
    error: directory.error,
    reportDialogVisible,
    defaultReportClassName,
    loadClasses: directory.loadClasses,
    handlePageChange: directory.handlePageChange,
    buildClassRoute,
    dashboardRoute: teacherDashboardRoute,
    openClassReportDialog,
  }
}
