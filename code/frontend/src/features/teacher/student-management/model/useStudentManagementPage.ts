import { computed, ref } from 'vue'

import { useAuthStore } from '@/stores/auth'
import {
  teacherClassManagementRoute,
  teacherStudentAnalysisRoute,
} from './teacherStudentManagementRoutes'
import { useTeacherStudentDirectory } from './useTeacherStudentDirectory'

export function useStudentManagementPage() {
  const authStore = useAuthStore()
  const reportDialogVisible = ref(false)
  const directory = useTeacherStudentDirectory()
  const classManagementRoute = computed(() => teacherClassManagementRoute(authStore.user?.role))

  function buildStudentRoute(studentId: string) {
    const student = directory.students.value.find((item) => item.id === studentId)

    return teacherStudentAnalysisRoute(
      authStore.user?.role,
      studentId,
      directory.selectedClassName.value || student?.class_name || ''
    )
  }

  function openClassReportDialog(): void {
    reportDialogVisible.value = true
  }

  return {
    classes: directory.classes,
    selectedClassName: directory.selectedClassName,
    searchQuery: directory.searchQuery,
    studentNoQuery: directory.studentNoQuery,
    students: directory.students,
    filteredTotal: directory.filteredTotal,
    totalStudents: directory.totalStudents,
    page: directory.page,
    totalPages: directory.totalPages,
    loadingClasses: directory.loadingClasses,
    loadingStudents: directory.loadingStudents,
    error: directory.error,
    classManagementRoute,
    reportDialogVisible,
    initialize: directory.initialize,
    openClassReportDialog,
    updateSearchQuery: directory.updateSearchQuery,
    updateStudentNoQuery: directory.updateStudentNoQuery,
    selectClass: directory.selectClass,
    handlePageChange: directory.handlePageChange,
    buildStudentRoute,
  }
}
