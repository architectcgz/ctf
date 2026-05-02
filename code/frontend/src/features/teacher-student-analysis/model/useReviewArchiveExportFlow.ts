import { ref, type Ref } from 'vue'

import { downloadReport } from '@/api/assessment'
import { ApiError } from '@/api/request'
import { exportStudentReviewArchive } from '@/api/teacher'
import type { ReportExportData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

interface UseReviewArchiveExportFlowOptions {
  selectedStudentId: Ref<string>
  startPolling: (
    reportId: string,
    onUpdate: (report: ReportExportData) => void,
    onError?: (error: unknown) => void
  ) => void
  stopPolling: () => void
}

export function useReviewArchiveExportFlow(options: UseReviewArchiveExportFlowOptions) {
  const { selectedStudentId, startPolling, stopPolling } = options
  const toast = useToast()

  const reviewArchiveSubmitting = ref(false)
  const downloadingReviewArchive = ref(false)
  const pendingReviewArchiveReportId = ref<string | null>(null)
  const reportDialogVisible = ref(false)

  function openClassReportDialog() {
    reportDialogVisible.value = true
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

  function notifyReviewArchiveActionError(error: unknown, fallback: string): void {
    console.error(fallback, error)
    if (error instanceof ApiError) {
      return
    }
    const message = error instanceof Error && error.message.trim() ? error.message : fallback
    toast.error(message)
  }

  async function downloadReviewArchiveReport(reportId: string): Promise<void> {
    try {
      await downloadGeneratedReport(reportId)
      toast.success('复盘归档已生成并开始下载')
    } catch (error) {
      notifyReviewArchiveActionError(error, '复盘归档下载失败，请稍后重试')
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
        pendingReviewArchiveReportId.value = null
        stopPolling()
        await downloadReviewArchiveReport(result.report_id)
        return
      }

      if (result.status === 'failed') {
        pendingReviewArchiveReportId.value = null
        stopPolling()
        toast.error(result.error_message || '复盘归档生成失败')
        return
      }

      pendingReviewArchiveReportId.value = result.report_id
      startPolling(
        result.report_id,
        (next) => {
          if (next.report_id !== pendingReviewArchiveReportId.value) return
          if (next.status === 'ready') {
            pendingReviewArchiveReportId.value = null
            stopPolling()
            void downloadReviewArchiveReport(next.report_id)
            return
          }
          if (next.status === 'failed') {
            pendingReviewArchiveReportId.value = null
            stopPolling()
            toast.error(next.error_message || '复盘归档生成失败')
          }
        },
        (error) => {
          pendingReviewArchiveReportId.value = null
          notifyReviewArchiveActionError(error, '复盘归档生成状态同步失败，请稍后重试')
        }
      )
      toast.info('复盘归档开始生成，完成后会自动下载')
    } catch (error) {
      pendingReviewArchiveReportId.value = null
      stopPolling()
      notifyReviewArchiveActionError(error, '复盘归档导出失败，请稍后重试')
    } finally {
      reviewArchiveSubmitting.value = false
    }
  }

  return {
    reportDialogVisible,
    reviewArchiveSubmitting,
    downloadingReviewArchive,
    pendingReviewArchiveReportId,
    openClassReportDialog,
    handleExportReviewArchive,
  }
}
