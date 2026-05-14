<script setup lang="ts">
import ClassStudentsPage from '@/components/teacher/class-management/ClassStudentsPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useTeacherClassStudentsPage } from '@/features/teacher-class-students'

const {
  selectedClassName,
  students,
  review,
  summary,
  trend,
  studentNoQuery,
  loadingStudents,
  error,
  reportDialogVisible,
  insightWindowFromDate,
  insightWindowToDate,
  insightWindowError,
  insightWindowLabel,
  activeInsightWindowFromDate,
  activeInsightWindowToDate,
  canApplyInsightWindow,
  canResetInsightWindow,
  initialize,
  openClassManagement,
  openDashboard,
  openClassReportDialog,
  updateStudentNoQuery,
  updateInsightWindowFromDate,
  updateInsightWindowToDate,
  applyInsightWindow,
  resetInsightWindow,
  openStudent,
} = useTeacherClassStudentsPage()
</script>

<template>
  <section class="teacher-route-root">
    <ClassStudentsPage
      :selected-class-name="selectedClassName"
      :students="students"
      :review="review"
      :summary="summary"
      :trend="trend"
      :student-no-query="studentNoQuery"
      :loading-students="loadingStudents"
      :error="error"
      :insight-window-from-date="insightWindowFromDate"
      :insight-window-to-date="insightWindowToDate"
      :insight-window-error="insightWindowError"
      :insight-window-label="insightWindowLabel"
      :can-apply-insight-window="canApplyInsightWindow"
      :can-reset-insight-window="canResetInsightWindow"
      @retry="initialize"
      @open-class-management="openClassManagement"
      @open-dashboard="openDashboard"
      @open-report-export="openClassReportDialog"
      @update-student-no-query="updateStudentNoQuery"
      @update-insight-window-from-date="updateInsightWindowFromDate"
      @update-insight-window-to-date="updateInsightWindowToDate"
      @apply-insight-window="applyInsightWindow"
      @reset-insight-window="resetInsightWindow"
      @open-student="openStudent"
    />
    <TeacherClassReportExportDialog
      v-model="reportDialogVisible"
      :default-class-name="selectedClassName"
      :default-from-date="activeInsightWindowFromDate"
      :default-to-date="activeInsightWindowToDate"
    />
  </section>
</template>
