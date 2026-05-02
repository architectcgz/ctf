<script setup lang="ts">
import ClassStudentsPage from '@/components/teacher/class-management/ClassStudentsPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useTeacherClassStudentsPage } from '@/features/teacher-class-students'

const {
  classes,
  selectedClassName,
  students,
  review,
  summary,
  trend,
  studentNoQuery,
  loadingStudents,
  error,
  reportDialogVisible,
  initialize,
  openClassManagement,
  openDashboard,
  openClassReportDialog,
  selectClass,
  updateStudentNoQuery,
  openStudent,
} = useTeacherClassStudentsPage()
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
    @open-class-management="openClassManagement"
    @open-dashboard="openDashboard"
    @open-report-export="openClassReportDialog"
    @select-class="selectClass"
    @update-student-no-query="updateStudentNoQuery"
    @open-student="openStudent"
  />
  <TeacherClassReportExportDialog
    v-model="reportDialogVisible"
    :default-class-name="selectedClassName"
  />
</template>
