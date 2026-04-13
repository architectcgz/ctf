<script setup lang="ts">
import { ref } from 'vue'

import StudentAnalysisPage from '@/components/teacher/class-management/StudentAnalysisPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useTeacherStudentAnalysisPage } from '@/composables/useTeacherStudentAnalysisPage'

const reportDialogVisible = ref(false)

const {
  router,
  classes,
  students,
  selectedClassName,
  selectedStudentId,
  selectedStudent,
  loadingClasses,
  loadingStudents,
  loadingDetails,
  error,
  progress,
  skillProfile,
  recommendations,
  timeline,
  evidence,
  writeupSubmissions,
  writeupPage,
  writeupTotal,
  writeupTotalPages,
  writeupPaginationLoading,
  manualReviewSubmissions,
  activeManualReview,
  manualReviewLoading,
  manualReviewSaving,
  solvedRate,
  weakDimensions,
  initialize,
  selectClass,
  selectStudent,
  openChallenge,
  openReviewArchivePage,
  handleExportReviewArchive,
  openManualReview,
  moderateWriteup,
  reviewManualReview,
  changeWriteupPage,
} = useTeacherStudentAnalysisPage()

function openClassReportDialog(): void {
  reportDialogVisible.value = true
}
</script>

<template>
  <StudentAnalysisPage
    :classes="classes"
    :students="students"
    :selected-class-name="selectedClassName"
    :selected-student-id="selectedStudentId"
    :selected-student="selectedStudent"
    :loading-classes="loadingClasses"
    :loading-students="loadingStudents"
    :loading-details="loadingDetails"
    :error="error"
    :progress="progress"
    :skill-profile="skillProfile"
    :recommendations="recommendations"
    :timeline="timeline"
    :evidence="evidence"
    :writeup-submissions="writeupSubmissions"
    :writeup-page="writeupPage"
    :writeup-total="writeupTotal"
    :writeup-total-pages="writeupTotalPages"
    :writeup-pagination-loading="writeupPaginationLoading"
    :manual-review-submissions="manualReviewSubmissions"
    :active-manual-review="activeManualReview"
    :manual-review-loading="manualReviewLoading"
    :manual-review-saving="manualReviewSaving"
    :solved-rate="solvedRate"
    :weak-dimensions="weakDimensions"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
    @open-class-students="router.push({ name: 'TeacherClassStudents', params: { className: selectedClassName } })"
    @open-report-export="openClassReportDialog"
    @open-review-archive="openReviewArchivePage"
    @export-review-archive="handleExportReviewArchive"
    @select-class="selectClass"
    @select-student="selectStudent"
    @open-challenge="openChallenge"
    @open-manual-review="openManualReview"
    @moderate-writeup="moderateWriteup"
    @review-manual-review="reviewManualReview"
    @change-writeup-page="changeWriteupPage"
  />
  <TeacherClassReportExportDialog
    v-model="reportDialogVisible"
    :default-class-name="selectedClassName"
  />
</template>
