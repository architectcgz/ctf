import { computed, ref, type Ref } from 'vue'

import { downloadReport } from '@/api/assessment'
import { ApiError } from '@/api/request'
import { exportStudentReviewArchive } from '@/api/teaching'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import { reportFrontendError } from '@/utils/reportFrontendError'
import {
  resolveStudentReviewArchiveErrorMessage,
  STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES,
  useStudentReviewArchive,
} from '@/features/teaching/student-review-archive'
import {
  studentReviewArchiveAnalysisRoute,
  studentReviewArchiveBackRoute,
} from './studentReviewArchiveRoutes'

interface UseStudentReviewArchivePageOptions {
  className: Readonly<Ref<string>>
  studentId: Readonly<Ref<string>>
}

export function useStudentReviewArchivePage({
  className,
  studentId,
}: UseStudentReviewArchivePageOptions) {
  const toast = useToast()
  const authStore = useAuthStore()
  const { start: startPolling, stop: stopPolling } = useReportStatusPolling()
  const { archive, loading, error, reload } = useStudentReviewArchive(studentId)

  const exporting = ref(false)
  const pendingReportId = ref<string | null>(null)
  const analysisRoute = computed(() =>
    studentReviewArchiveAnalysisRoute(authStore.user?.role, className.value, studentId.value)
  )
  const backRoute = computed(() =>
    studentReviewArchiveBackRoute(authStore.user?.role, className.value)
  )

  async function downloadGeneratedReport(reportId: string): Promise<void> {
    const { blob, filename } = await downloadReport(reportId)
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(objectUrl)
  }

  function notifyExportActionError(error: unknown, fallback: string): void {
    reportFrontendError(fallback, error)
    if (error instanceof ApiError) {
      return
    }
    toast.error(resolveStudentReviewArchiveErrorMessage(error, fallback))
  }

  async function downloadArchiveReport(reportId: string): Promise<void> {
    try {
      await downloadGeneratedReport(reportId)
      toast.success(STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.success)
    } catch (error) {
      notifyExportActionError(error, STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.downloadFailed)
    }
  }

  async function exportArchive(): Promise<void> {
    if (!studentId.value) return

    exporting.value = true
    try {
      const result = await exportStudentReviewArchive(studentId.value, { format: 'json' })
      if (result.status === 'ready') {
        pendingReportId.value = null
        stopPolling()
        await downloadArchiveReport(result.report_id)
        return
      }
      if (result.status === 'failed') {
        pendingReportId.value = null
        stopPolling()
        toast.error(
          result.error_message || STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.generationFailed
        )
        return
      }

      pendingReportId.value = result.report_id
      startPolling(
        result.report_id,
        (next) => {
          if (next.report_id !== pendingReportId.value) return
          if (next.status === 'ready') {
            pendingReportId.value = null
            stopPolling()
            void downloadArchiveReport(next.report_id)
            return
          }
          if (next.status === 'failed') {
            pendingReportId.value = null
            stopPolling()
            toast.error(
              next.error_message || STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.generationFailed
            )
          }
        },
        (error) => {
          pendingReportId.value = null
          notifyExportActionError(
            error,
            STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.pollingFailed
          )
        }
      )
      toast.info(STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.pending)
    } catch (error) {
      pendingReportId.value = null
      stopPolling()
      notifyExportActionError(error, STUDENT_REVIEW_ARCHIVE_EXPORT_MESSAGES.exportFailed)
    } finally {
      exporting.value = false
    }
  }

  return {
    archive,
    loading,
    error,
    reload,
    exporting,
    analysisRoute,
    backRoute,
    exportArchive,
  }
}
