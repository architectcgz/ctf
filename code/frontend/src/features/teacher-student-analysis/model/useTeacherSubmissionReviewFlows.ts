import { ref, type Ref } from 'vue'

import {
  getTeacherManualReviewSubmission,
  getTeacherManualReviewSubmissions,
  getTeacherWriteupSubmissions,
  hideTeacherCommunityWriteup,
  recommendTeacherCommunityWriteup,
  restoreTeacherCommunityWriteup,
  reviewTeacherManualReviewSubmission,
  unrecommendTeacherCommunityWriteup,
} from '@/api/teacher'
import type {
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseTeacherSubmissionReviewFlowsOptions {
  getCurrentStudentId: () => string
}

export function useTeacherSubmissionReviewFlows(options: UseTeacherSubmissionReviewFlowsOptions) {
  const { getCurrentStudentId } = options
  const toast = useToast()

  const writeupSubmissions = ref<TeacherSubmissionWriteupItemData[]>([])
  const writeupPage = ref(1)
  const writeupPageSize = ref(6)
  const writeupTotal = ref(0)
  const writeupPaginationLoading = ref(false)
  const manualReviewSubmissions = ref<TeacherManualReviewSubmissionItemData[]>([])
  const activeManualReview = ref<TeacherManualReviewSubmissionDetailData | null>(null)
  const manualReviewLoading = ref(false)
  const manualReviewSaving = ref(false)

  function resetSubmissionReviewState() {
    writeupSubmissions.value = []
    writeupPage.value = 1
    writeupTotal.value = 0
    manualReviewSubmissions.value = []
    activeManualReview.value = null
  }

  function applyWriteupPagePayload(payload: {
    list: TeacherSubmissionWriteupItemData[]
    page: number
    page_size: number
    total: number
  }) {
    writeupSubmissions.value = payload.list
    writeupPage.value = payload.page
    writeupPageSize.value = payload.page_size
    writeupTotal.value = payload.total
  }

  async function refreshWriteupSubmissions(studentId = getCurrentStudentId(), targetPage = writeupPage.value) {
    if (!studentId) {
      writeupSubmissions.value = []
      writeupPage.value = 1
      writeupTotal.value = 0
      return
    }
    writeupPaginationLoading.value = true
    try {
      const nextWriteups = await getTeacherWriteupSubmissions({
        student_id: studentId,
        submission_status: 'published',
        page: targetPage,
        page_size: writeupPageSize.value,
      })
      const totalPages = Math.max(
        1,
        Math.ceil(nextWriteups.total / Math.max(1, nextWriteups.page_size))
      )
      if (targetPage > totalPages) {
        writeupPaginationLoading.value = false
        await refreshWriteupSubmissions(studentId, totalPages)
        return
      }
      applyWriteupPagePayload(nextWriteups)
    } finally {
      writeupPaginationLoading.value = false
    }
  }

  async function changeWriteupPage(page: number): Promise<void> {
    if (page < 1 || page === writeupPage.value || writeupPaginationLoading.value) return
    await refreshWriteupSubmissions(getCurrentStudentId(), page)
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
      const currentStudentId = getCurrentStudentId()
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

  async function moderateWriteup(payload: {
    submissionId: string
    action: 'recommend' | 'unrecommend' | 'hide' | 'restore'
  }): Promise<void> {
    switch (payload.action) {
      case 'recommend':
        await recommendTeacherCommunityWriteup(payload.submissionId)
        toast.success('已设为推荐题解')
        break
      case 'unrecommend':
        await unrecommendTeacherCommunityWriteup(payload.submissionId)
        toast.success('已取消推荐题解')
        break
      case 'hide':
        await hideTeacherCommunityWriteup(payload.submissionId)
        toast.success('已隐藏社区题解')
        break
      case 'restore':
        await restoreTeacherCommunityWriteup(payload.submissionId)
        toast.success('已恢复社区题解')
        break
    }
    await refreshWriteupSubmissions()
  }

  return {
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
  }
}
