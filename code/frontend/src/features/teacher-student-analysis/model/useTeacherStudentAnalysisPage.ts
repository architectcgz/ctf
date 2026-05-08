import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  getClasses,
  getClassStudents,
  getStudentProgress,
  getStudentRecommendations,
  getStudentSkillProfile,
  getStudentTimeline,
  getTeacherManualReviewSubmissions,
  getTeacherWriteupSubmissions,
} from '@/api/teacher'
import type {
  MyProgressData,
  RecommendationItem,
  RecommendationWeakDimension,
  SkillProfileData,
  TeacherClassItem,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useAuthStore } from '@/stores/auth'
import { getWeakDimensionLabels } from '@/utils/skillProfile'
import { useReviewArchiveExportFlow } from './useReviewArchiveExportFlow'
import { useTeacherReviewWorkspace } from './useTeacherReviewWorkspace'
import { useTeacherStudentAnalysisNavigation } from './useTeacherStudentAnalysisNavigation'
import { useTeacherSubmissionReviewFlows } from './useTeacherSubmissionReviewFlows'

export function useTeacherStudentAnalysisPage() {
  const route = useRoute()
  const router = useRouter()
  const authStore = useAuthStore()
  const { start: startPolling, stop: stopPolling } = useReportStatusPolling()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const classes = ref<TeacherClassItem[]>([])
  const students = ref<TeacherStudentItem[]>([])
  const selectedClassName = ref('')
  const selectedStudentId = ref('')

  const loadingClasses = ref(false)
  const loadingStudents = ref(false)
  const loadingDetails = ref(false)
  const error = ref<string | null>(null)

  const progress = ref<MyProgressData | null>(null)
  const skillProfile = ref<SkillProfileData | null>(null)
  const recommendations = ref<RecommendationItem[]>([])
  const weakDimensionAdvice = ref<RecommendationWeakDimension[]>([])
  const timeline = ref<TimelineEvent[]>([])
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
  } = useTeacherReviewWorkspace()
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
    applyWriteupPagePayload,
    refreshWriteupSubmissions,
    changeWriteupPage,
    openManualReview,
    reviewManualReview,
    moderateWriteup,
  } = useTeacherSubmissionReviewFlows({
    getCurrentStudentId: studentIdFromRoute,
  })

  const selectedStudent = computed(
    () => students.value.find((item) => item.id === selectedStudentId.value) ?? null
  )
  const solvedRate = computed(() => {
    if (!progress.value?.total_challenges) return 0
    return Math.round(
      ((progress.value.solved_challenges ?? 0) / progress.value.total_challenges) * 100
    )
  })
  const weakDimensions = computed(() => getWeakDimensionLabels(weakDimensionAdvice.value))
  const writeupTotalPages = computed(() =>
    Math.max(1, Math.ceil(writeupTotal.value / Math.max(1, writeupPageSize.value)))
  )
  const {
    reportDialogVisible,
    reviewArchiveSubmitting,
    downloadingReviewArchive,
    pendingReviewArchiveReportId,
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

  function syncReviewWorkspaceQueryFromRoute(): void {
    const nextQuery = reviewWorkspaceQueryFromRoute()

    setSessionQuery({
      with_events: true,
      limit: 20,
      offset: 0,
      ...nextQuery,
    })
  }

  function reviewWorkspaceQueryFromRoute(): Partial<typeof sessionQuery.value> {
    return {
      mode:
        route.query.reviewMode === 'practice' ||
        route.query.reviewMode === 'jeopardy' ||
        route.query.reviewMode === 'awd'
          ? route.query.reviewMode
          : undefined,
      result:
        route.query.reviewResult === 'success' ||
        route.query.reviewResult === 'failed' ||
        route.query.reviewResult === 'in_progress' ||
        route.query.reviewResult === 'unknown'
          ? route.query.reviewResult
          : undefined,
      challenge_id:
        typeof route.query.reviewChallengeId === 'string' && route.query.reviewChallengeId.trim()
          ? route.query.reviewChallengeId.trim()
          : undefined,
    }
  }

  function reviewWorkspaceQueryMatchesState(
    nextQuery: Partial<typeof sessionQuery.value>
  ): boolean {
    return (
      (nextQuery.mode || undefined) === (sessionQuery.value.mode || undefined) &&
      (nextQuery.result || undefined) === (sessionQuery.value.result || undefined) &&
      (nextQuery.challenge_id || undefined) === (sessionQuery.value.challenge_id || undefined)
    )
  }

  async function loadClasses(): Promise<void> {
    loadingClasses.value = true
    try {
      classes.value = await getClasses()
    } finally {
      loadingClasses.value = false
    }
  }

  async function loadStudents(className = classNameFromRoute()): Promise<void> {
    if (!className) {
      selectedClassName.value = ''
      students.value = []
      return
    }

    loadingStudents.value = true
    selectedClassName.value = className

    try {
      students.value = await getClassStudents(className)
    } finally {
      loadingStudents.value = false
    }
  }

  async function loadStudentDetails(studentId = studentIdFromRoute()): Promise<void> {
    if (!studentId) {
      progress.value = null
      skillProfile.value = null
      recommendations.value = []
      weakDimensionAdvice.value = []
      timeline.value = []
      resetReviewWorkspace()
      resetSubmissionReviewState()
      selectedStudentId.value = ''
      return
    }

    loadingDetails.value = true
    selectedStudentId.value = studentId

    try {
      const [
        nextProgress,
        nextProfile,
        nextRecommendations,
        nextTimeline,
        _reviewWorkspaceLoaded,
        nextWriteups,
        nextManualReviews,
      ] = await Promise.all([
        getStudentProgress(studentId),
        getStudentSkillProfile(studentId),
        getStudentRecommendations(studentId),
        getStudentTimeline(studentId),
        loadReviewWorkspace(studentId),
        getTeacherWriteupSubmissions({
          student_id: studentId,
          submission_status: 'published',
          page: writeupPage.value,
          page_size: writeupPageSize.value,
        }),
        getTeacherManualReviewSubmissions({ student_id: studentId, page_size: 6 }),
      ])

      progress.value = nextProgress
      skillProfile.value = nextProfile
      recommendations.value = nextRecommendations.challenges
      weakDimensionAdvice.value = nextRecommendations.weak_dimensions
      timeline.value = nextTimeline
      void _reviewWorkspaceLoaded
      applyWriteupPagePayload(nextWriteups)
      manualReviewSubmissions.value = nextManualReviews.list
      activeManualReview.value = null
    } finally {
      loadingDetails.value = false
    }
  }

  async function initialize(): Promise<void> {
    error.value = null

    try {
      syncReviewWorkspaceQueryFromRoute()
      await loadClasses()
      await loadStudents()
      await loadStudentDetails()
    } catch (err) {
      console.error('加载学员分析失败:', err)
      error.value = '加载学员分析失败，请稍后重试'
    }
  }

  async function updateReviewWorkspaceFilters(
    nextQuery: Partial<{
      challenge_id: string
      mode: 'practice' | 'jeopardy' | 'awd'
      result: 'success' | 'failed' | 'in_progress' | 'unknown'
    }>
  ): Promise<void> {
    const studentId = selectedStudentId.value || studentIdFromRoute()
    const mergedQuery = {
      ...sessionQuery.value,
      ...nextQuery,
      offset: 0,
    }

    setSessionQuery(mergedQuery)

    await router.replace({
      query: {
        ...route.query,
        reviewMode: mergedQuery.mode || undefined,
        reviewResult: mergedQuery.result || undefined,
        reviewChallengeId: mergedQuery.challenge_id || undefined,
      },
    })

    if (Object.prototype.hasOwnProperty.call(nextQuery, 'challenge_id')) {
      await loadReviewWorkspace(studentId)
      return
    }

    await reloadAttackSessions(studentId)
  }
  const {
    selectClass,
    openClassManagement,
    openClassStudents,
    selectStudent,
    openChallenge,
    openReviewArchivePage,
  } = useTeacherStudentAnalysisNavigation({
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
    openClassManagement,
    openClassStudents,
    selectClass,
    selectStudent,
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
