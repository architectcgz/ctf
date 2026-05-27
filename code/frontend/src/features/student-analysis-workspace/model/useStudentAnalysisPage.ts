import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useAuthStore } from '@/stores/auth'
import { reportFrontendError } from '@/utils/reportFrontendError'
import {
  useReviewArchiveExportFlow,
  useReviewWorkspace,
  useSubmissionReviewFlows,
} from '@/features/student-analysis-review'
import { useStudentAnalysisDataState } from './useStudentAnalysisDataState'
import { useStudentAnalysisNavigation } from './useStudentAnalysisNavigation'
import { useStudentAnalysisReviewQuerySync } from './useStudentAnalysisReviewQuerySync'

export function useStudentAnalysisPage() {
  const route = useRoute()
  const router = useRouter()
  const authStore = useAuthStore()
  const { start: startPolling, stop: stopPolling } = useReportStatusPolling()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const error = ref<string | null>(null)
  const {
    evidence,
    attackSessions,
    reviewChallengeOptions,
    reviewWorkspaceLoading,
    sessionQuery,
    loadReviewWorkspace,
    reloadAttackSessions,
    resetReviewWorkspace,
    setSessionQuery,
  } = useReviewWorkspace()
  const {
    writeupSubmissions,
    writeupPage,
    writeupPageSize,
    writeupTotal,
    writeupPaginationLoading,
    manualReviewSubmissions,
    activeManualReview,
    manualReviewLoading,
    manualReviewSaving,
    resetSubmissionReviewState,
    refreshWriteupSubmissions,
    refreshManualReviewSubmissions,
    changeWriteupPage,
    openManualReview,
    reviewManualReview,
    moderateWriteup,
  } = useSubmissionReviewFlows({
    getCurrentStudentId: studentIdFromRoute,
  })

  const {
    selectedClassName,
    selectedStudentId,
    selectedStudent,
    loadingDetails,
    progress,
    skillProfile,
    recommendations,
    timeline,
    solvedRate,
    weakDimensions,
    loadStudents,
    loadStudentDetails,
  } = useStudentAnalysisDataState({
    classNameFromRoute,
    studentIdFromRoute,
  })
  const writeupTotalPages = computed(() =>
    Math.max(1, Math.ceil(writeupTotal.value / Math.max(1, writeupPageSize.value)))
  )
  const {
    reportDialogVisible,
    openClassReportDialog,
    handleExportReviewArchive,
  } = useReviewArchiveExportFlow({
    selectedStudentId,
    startPolling,
    stopPolling,
  })

  function classNameFromRoute(): string {
    return String(route.params.className || '')
  }

  function studentIdFromRoute(): string {
    return String(route.params.studentId || '')
  }

  const {
    reviewWorkspaceQueryFromRoute,
    syncReviewWorkspaceQueryFromRoute,
    reviewWorkspaceQueryMatchesState,
    updateReviewWorkspaceFilters,
  } = useStudentAnalysisReviewQuerySync({
    route,
    router,
    sessionQuery,
    selectedStudentId,
    setSessionQuery,
    loadReviewWorkspace,
    reloadAttackSessions,
    studentIdFromRoute,
  })

  async function initialize(): Promise<void> {
    error.value = null

    try {
      syncReviewWorkspaceQueryFromRoute()
      await loadStudents()
      const studentId = studentIdFromRoute()

      if (!studentId) {
        await loadStudentDetails('')
        resetReviewWorkspace()
        resetSubmissionReviewState()
        return
      }

      await Promise.all([
        loadStudentDetails(studentId),
        loadReviewWorkspace(studentId),
        refreshWriteupSubmissions(studentId),
        refreshManualReviewSubmissions(studentId),
      ])
    } catch (err) {
      reportFrontendError('加载学员分析失败:', err)
      error.value = '加载学员分析失败，请稍后重试'
    }
  }

  const {
    openClassStudents,
    openChallenge,
    openReviewArchivePage,
  } = useStudentAnalysisNavigation({
    router,
    getRole: () => authStore.user?.role,
    selectedClassName,
    selectedStudentId,
  })

  watch(
    () => [route.params.className, route.params.studentId],
    () => {
      void initialize()
    }
  )

  watch(
    () => [route.query.reviewMode, route.query.reviewResult, route.query.reviewChallengeId],
    () => {
      const studentId = selectedStudentId.value || studentIdFromRoute()
      if (!studentId) return

      const nextQuery = reviewWorkspaceQueryFromRoute()
      if (reviewWorkspaceQueryMatchesState(nextQuery)) {
        return
      }
      const challengeChanged =
        (nextQuery.challenge_id || undefined) !== (sessionQuery.value.challenge_id || undefined)

      syncReviewWorkspaceQueryFromRoute()
      if (challengeChanged) {
        void loadReviewWorkspace(studentId)
        return
      }
      void reloadAttackSessions(studentId)
    }
  )

  watch(
    () => [selectedStudent.value?.name, selectedStudent.value?.username, selectedStudentId.value],
    () => {
      setBreadcrumbDetailTitle(
        selectedStudent.value?.name || selectedStudent.value?.username || selectedStudentId.value
      )
    },
    { immediate: true }
  )

  onMounted(() => {
    void initialize()
  })

  onUnmounted(() => {
    setBreadcrumbDetailTitle()
  })

  return {
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
    reviewWorkspaceQuery: sessionQuery,
    writeupSubmissions,
    writeupPage,
    writeupPageSize,
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
  }
}
