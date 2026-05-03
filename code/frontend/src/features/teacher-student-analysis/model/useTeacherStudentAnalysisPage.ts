import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  getClasses,
  getClassStudents,
  getStudentEvidence,
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
  SkillProfileData,
  TeacherClassItem,
  TeacherEvidenceData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useAuthStore } from '@/stores/auth'
import { getWeakDimensions } from '@/utils/skillProfile'
import { useReviewArchiveExportFlow } from './useReviewArchiveExportFlow'
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
  const timeline = ref<TimelineEvent[]>([])
  const evidence = ref<TeacherEvidenceData | null>(null)
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
  const weakDimensions = computed(() => getWeakDimensions(skillProfile.value))
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
      timeline.value = []
      evidence.value = null
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
        nextEvidence,
        nextWriteups,
        nextManualReviews,
      ] = await Promise.all([
        getStudentProgress(studentId),
        getStudentSkillProfile(studentId),
        getStudentRecommendations(studentId),
        getStudentTimeline(studentId),
        getStudentEvidence(studentId),
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
      recommendations.value = nextRecommendations
      timeline.value = nextTimeline
      evidence.value = nextEvidence
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
      await loadClasses()
      await loadStudents()
      await loadStudentDetails()
    } catch (err) {
      console.error('加载学员分析失败:', err)
      error.value = '加载学员分析失败，请稍后重试'
    }
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
  }
}
