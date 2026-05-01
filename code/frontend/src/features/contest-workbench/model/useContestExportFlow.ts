import { ref } from 'vue'

import { downloadReport } from '@/api/assessment'
import { exportContestArchive } from '@/api/admin/contests'
import type { ContestDetailData } from '@/api/contracts'
import { ApiError } from '@/api/request'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useToast } from '@/composables/useToast'

export function useContestExportFlow() {
  const toast = useToast()
  const { start: startPolling, stop: stopPolling } = useReportStatusPolling()

  const exportingContestId = ref<string | null>(null)
  const downloadingContestReport = ref(false)
  const pendingContestReportId = ref<string | null>(null)

  async function downloadGeneratedReport(reportId: string): Promise<void> {
    downloadingContestReport.value = true
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
      downloadingContestReport.value = false
    }
  }

  function notifyContestExportError(error: unknown, fallback: string): void {
    console.error(fallback, error)
    if (error instanceof ApiError) {
      return
    }
    const message = error instanceof Error && error.message.trim() ? error.message : fallback
    toast.error(message)
  }

  async function downloadContestReport(reportId: string, contestTitle: string): Promise<void> {
    try {
      await downloadGeneratedReport(reportId)
      toast.success(`赛事结果已导出：${contestTitle}`)
    } catch (error) {
      notifyContestExportError(error, `赛事结果下载失败：${contestTitle}`)
    }
  }

  async function handleExportContest(contest: ContestDetailData): Promise<void> {
    exportingContestId.value = contest.id
    try {
      const result = await exportContestArchive(contest.id, { format: 'json' })

      if (result.status === 'ready') {
        pendingContestReportId.value = null
        stopPolling()
        await downloadContestReport(result.report_id, contest.title)
        return
      }

      if (result.status === 'failed') {
        pendingContestReportId.value = null
        stopPolling()
        toast.error(result.error_message || '赛事结果导出失败')
        return
      }

      pendingContestReportId.value = result.report_id
      startPolling(
        result.report_id,
        (next) => {
          if (next.report_id !== pendingContestReportId.value) return
          if (next.status === 'ready') {
            pendingContestReportId.value = null
            stopPolling()
            void downloadContestReport(next.report_id, contest.title)
            return
          }
          if (next.status === 'failed') {
            pendingContestReportId.value = null
            stopPolling()
            toast.error(next.error_message || '赛事结果导出失败')
          }
        },
        (error) => {
          pendingContestReportId.value = null
          notifyContestExportError(error, `赛事结果生成状态同步失败：${contest.title}`)
        }
      )
      toast.info(`已开始导出赛事结果：${contest.title}`)
    } catch (error) {
      pendingContestReportId.value = null
      stopPolling()
      notifyContestExportError(error, `赛事结果导出失败：${contest.title}`)
    } finally {
      exportingContestId.value = null
    }
  }

  return {
    handleExportContest,
  }
}
