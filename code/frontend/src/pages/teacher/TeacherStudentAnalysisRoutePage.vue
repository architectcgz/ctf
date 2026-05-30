<template>
  <section class="teacher-route-root">
    <StudentAnalysisPage
      :selected-student="selectedStudent"
      :loading-details="loadingDetails"
      :error="error"
      :progress="progress"
      :skill-profile="skillProfile"
      :recommendations="recommendations"
      :timeline="timeline"
      :evidence="evidence"
      :attack-sessions="attackSessions"
      :review-challenge-options="reviewChallengeOptions"
      :review-workspace-loading="reviewWorkspaceLoading"
      :review-workspace-query="reviewWorkspaceQuery"
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
      @open-class-students="openClassStudents"
      @open-report-export="openClassReportDialog"
      @open-review-archive="openReviewArchivePage"
      @export-review-archive="handleExportReviewArchive"
      @open-challenge="openChallenge"
      @open-manual-review="openManualReview"
      @moderate-writeup="moderateWriteup"
      @review-manual-review="reviewManualReview"
      @change-writeup-page="changeWriteupPage"
      @update-review-workspace-filters="updateReviewWorkspaceFilters"
    />
    <ClassReportExportDialog
      v-model="reportDialogVisible"
      :default-class-name="selectedClassName"
    />
  </section>
</template>

<script setup lang="ts">
import { ClassReportExportDialog } from '@/features/teaching/class-report-export'
import {
  StudentAnalysisPage,
  useStudentAnalysisPage,
} from '@/features/teaching/student-analysis-workspace'

const {
  selectedClassName,
  selectedStudent,
  loadingDetails,
  error,
  progress,
  skillProfile,
  recommendations,
  timeline,
  evidence,
  attackSessions,
  reviewChallengeOptions,
  reviewWorkspaceLoading,
  reviewWorkspaceQuery,
  writeupSubmissions,
  writeupPage,
  writeupTotal,
  writeupTotalPages,
  writeupPaginationLoading,
  manualReviewSubmissions,
  activeManualReview,
  manualReviewLoading,
  manualReviewSaving,
  reportDialogVisible,
  solvedRate,
  weakDimensions,
  initialize,
  openClassStudents,
  openChallenge,
  openClassReportDialog,
  openReviewArchivePage,
  handleExportReviewArchive,
  openManualReview,
  moderateWriteup,
  reviewManualReview,
  changeWriteupPage,
  updateReviewWorkspaceFilters,
} = useStudentAnalysisPage()
</script>
