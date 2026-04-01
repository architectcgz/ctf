<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { downloadReport } from '@/api/assessment'
import {
  exportStudentReviewArchive,
  getClasses,
  getClassStudents,
  getTeacherManualReviewSubmission,
  getTeacherManualReviewSubmissions,
  getStudentEvidence,
  getStudentProgress,
  getStudentRecommendations,
  getStudentSkillProfile,
  getStudentTimeline,
  getTeacherWriteupSubmissions,
  reviewTeacherManualReviewSubmission,
} from '@/api/teacher'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherEvidenceData,
  TeacherClassItem,
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import StudentAnalysisPage from '@/components/teacher/class-management/StudentAnalysisPage.vue'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useToast } from '@/composables/useToast'
import { getWeakDimensions } from '@/utils/skillProfile'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { start: startPolling, stop: stopPolling } = useReportStatusPolling()

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
const writeupSubmissions = ref<TeacherSubmissionWriteupItemData[]>([])
const manualReviewSubmissions = ref<TeacherManualReviewSubmissionItemData[]>([])
const activeManualReview = ref<TeacherManualReviewSubmissionDetailData | null>(null)
const manualReviewLoading = ref(false)
const manualReviewSaving = ref(false)
const reviewArchiveSubmitting = ref(false)
const downloadingReviewArchive = ref(false)
const pendingReviewArchiveReportId = ref<string | null>(null)

const selectedStudent = computed(() => students.value.find((item) => item.id === selectedStudentId.value) ?? null)
const solvedRate = computed(() => {
  if (!progress.value?.total_challenges) return 0
  return Math.round(((progress.value.solved_challenges ?? 0) / progress.value.total_challenges) * 100)
})
const weakDimensions = computed(() => getWeakDimensions(skillProfile.value))

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
    writeupSubmissions.value = []
    manualReviewSubmissions.value = []
    activeManualReview.value = null
    selectedStudentId.value = ''
    return
  }

  loadingDetails.value = true
  selectedStudentId.value = studentId

  try {
    const [nextProgress, nextProfile, nextRecommendations, nextTimeline, nextEvidence, nextWriteups, nextManualReviews] = await Promise.all([
      getStudentProgress(studentId),
      getStudentSkillProfile(studentId),
      getStudentRecommendations(studentId),
      getStudentTimeline(studentId),
      getStudentEvidence(studentId),
      getTeacherWriteupSubmissions({ student_id: studentId, page_size: 6 }),
      getTeacherManualReviewSubmissions({ student_id: studentId, page_size: 6 }),
    ])

    progress.value = nextProgress
    skillProfile.value = nextProfile
    recommendations.value = nextRecommendations
    timeline.value = nextTimeline
    evidence.value = nextEvidence
    writeupSubmissions.value = nextWriteups.list
    manualReviewSubmissions.value = nextManualReviews.list
    activeManualReview.value = null
  } finally {
    loadingDetails.value = false
  }
}

async function openManualReview(submissionId: string): Promise<void> {
  manualReviewLoading.value = true
  try {
    activeManualReview.value = await getTeacherManualReviewSubmission(submissionId)
  } finally {
    manualReviewLoading.value = false
  }
}

async function reviewManualReview(payload: {
  submissionId: string
  reviewStatus: 'approved' | 'rejected'
  reviewComment?: string
}): Promise<void> {
  manualReviewSaving.value = true
  try {
    activeManualReview.value = await reviewTeacherManualReviewSubmission(payload.submissionId, {
      review_status: payload.reviewStatus,
      review_comment: payload.reviewComment,
    })
    const currentStudentId = studentIdFromRoute()
    if (currentStudentId) {
      const nextManualReviews = await getTeacherManualReviewSubmissions({
        student_id: currentStudentId,
        page_size: 6,
      })
      manualReviewSubmissions.value = nextManualReviews.list
    }
  } finally {
    manualReviewSaving.value = false
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

function selectClass(className: string): void {
  router.push({
    name: 'TeacherClassStudents',
    params: { className },
  })
}

function selectStudent(studentId: string): void {
  router.push({
    name: 'TeacherStudentAnalysis',
    params: {
      className: selectedClassName.value,
      studentId,
    },
  })
}

function openChallenge(challengeId: string): void {
  router.push(`/challenges/${challengeId}`)
}

function openReviewArchivePage(): void {
  if (!selectedStudentId.value || !selectedClassName.value) return
  router.push({
    name: 'TeacherStudentReviewArchive',
    params: {
      className: selectedClassName.value,
      studentId: selectedStudentId.value,
    },
  })
}

async function downloadGeneratedReport(reportId: string): Promise<void> {
  downloadingReviewArchive.value = true
  try {
    const { blob, filename } = await downloadReport(reportId)
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(objectUrl)
  } finally {
    downloadingReviewArchive.value = false
  }
}

async function handleExportReviewArchive(): Promise<void> {
  if (!selectedStudentId.value) {
    toast.warning('请先选择学生')
    return
  }

  reviewArchiveSubmitting.value = true
  try {
    const result = await exportStudentReviewArchive(selectedStudentId.value, { format: 'json' })

    if (result.status === 'ready') {
      stopPolling()
      await downloadGeneratedReport(result.report_id)
      toast.success('复盘归档已生成并开始下载')
      return
    }

    if (result.status === 'failed') {
      stopPolling()
      toast.error(result.error_message || '复盘归档生成失败')
      return
    }

    pendingReviewArchiveReportId.value = result.report_id
    startPolling(result.report_id, (next) => {
      if (next.report_id !== pendingReviewArchiveReportId.value) return
      if (next.status === 'ready') {
        pendingReviewArchiveReportId.value = null
        void downloadGeneratedReport(next.report_id)
        toast.success('复盘归档已生成并开始下载')
        return
      }
      if (next.status === 'failed') {
        pendingReviewArchiveReportId.value = null
        toast.error(next.error_message || '复盘归档生成失败')
      }
    })
    toast.info('复盘归档开始生成，完成后会自动下载')
  } finally {
    reviewArchiveSubmitting.value = false
  }
}

watch(
  () => [route.params.className, route.params.studentId],
  () => {
    initialize()
  },
)

onMounted(() => {
  initialize()
})
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
    :manual-review-submissions="manualReviewSubmissions"
    :active-manual-review="activeManualReview"
    :manual-review-loading="manualReviewLoading"
    :manual-review-saving="manualReviewSaving"
    :solved-rate="solvedRate"
    :weak-dimensions="weakDimensions"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
    @open-class-students="router.push({ name: 'TeacherClassStudents', params: { className: selectedClassName } })"
    @open-report-export="router.push({ name: 'ReportExport' })"
    @open-review-archive="openReviewArchivePage"
    @export-review-archive="handleExportReviewArchive"
    @select-class="selectClass"
    @select-student="selectStudent"
    @open-challenge="openChallenge"
    @open-manual-review="openManualReview"
    @review-manual-review="reviewManualReview"
  />
</template>
