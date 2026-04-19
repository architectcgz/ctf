<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import ClassStudentsPage from '@/components/teacher/class-management/ClassStudentsPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useTeacherClassWorkspacePage } from '@/composables/useTeacherClassWorkspacePage'
import { useAuthStore } from '@/stores/auth'
import {
  resolveClassManagementRouteName,
  resolveClassWorkspaceSectionRouteName,
  resolveClassStudentsRouteName,
  resolveStudentAnalysisRouteName,
  resolveTeachingDashboardRouteName,
} from '@/utils/teachingWorkspaceRouting'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const reportDialogVisible = ref(false)
const {
  classes,
  review,
  summary,
  trend,
  selectedClassName,
  students,
  studentNoQuery,
  loadingStudents,
  error,
  updateStudentNoQuery,
  initialize,
} = useTeacherClassWorkspacePage()

function pushClassRoute(routeName: string, className: string): void {
  if (!className) {
    return
  }

  if (className === selectedClassName.value && route.name === routeName) {
    return
  }

  const { panel: _panel, ...nextQuery } = route.query
  router.push({
    name: routeName,
    params: { className },
    query: nextQuery,
  })
}

function selectClass(className: string): void {
  pushClassRoute(resolveClassStudentsRouteName(authStore.user?.role), className)
}

function openWorkspaceSection(section: 'trend' | 'review' | 'insights' | 'intervention'): void {
  pushClassRoute(resolveClassWorkspaceSectionRouteName(authStore.user?.role, section), selectedClassName.value)
}

function openStudent(studentId: string): void {
  router.push({
    name: resolveStudentAnalysisRouteName(authStore.user?.role),
    params: {
      className: selectedClassName.value,
      studentId,
    },
  })
}

function openClassReportDialog(): void {
  reportDialogVisible.value = true
}
</script>

<template>
  <ClassStudentsPage
    :classes="classes"
    :selected-class-name="selectedClassName"
    :students="students"
    :review="review"
    :summary="summary"
    :trend="trend"
    :student-no-query="studentNoQuery"
    :loading-students="loadingStudents"
    :error="error"
    @retry="initialize"
    @open-class-management="
      router.push({ name: resolveClassManagementRouteName(authStore.user?.role) })
    "
    @open-dashboard="router.push({ name: resolveTeachingDashboardRouteName(authStore.user?.role) })"
    @open-report-export="openClassReportDialog"
    @open-workspace-section="openWorkspaceSection"
    @select-class="selectClass"
    @update-student-no-query="updateStudentNoQuery"
    @open-student="openStudent"
  />
  <TeacherClassReportExportDialog
    v-model="reportDialogVisible"
    :default-class-name="selectedClassName"
  />
</template>
