import { computed, ref, type ComputedRef, type Ref } from 'vue'

import { downloadReport } from '@/api/assessment'
import { ApiError } from '@/api/request'
import {
  exportTeacherAWDReviewArchive,
  exportTeacherAWDReviewReport,
} from '@/api/teacher'
import type { ReportExportData } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

type ExportKind = 'archive' | 'report'

interface UseTeacherAwdReviewExportFlowOptions {
  contestId: ComputedRef<string>
  selectedRoundNumber: ComputedRef<number | undefined>
  canExportReport: ComputedRef<boolean>
  startPolling: (
    reportId: string,
    onUpdate: (report: ReportExportData) => void,
    onError?: (error: unknown) => void
  ) => void
  stopPolling: () => void
}

export function useTeacherAwdReviewExportFlow(options: UseTeacherAwdReviewExportFlowOptions) {
  const { contestId, selectedRoundNumber, canExportReport, startPolling, stopPolling } = options
  const toast = useToast()

  const exporting = ref<ExportKind | null>(null)
  const pendingReportId = ref<string | null>(null)
  const polling = computed(() => Boolean(pendingReportId.value))

  function buildExportPayload(): { round_number?: number } | undefined {
    if (!selectedRoundNumber.value) return undefined
    return {
      round_number: selectedRoundNumber.value,
    }
  }

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

  function notifyExportError(error: unknown, fallback: string): void {
    console.error(fallback, error)
    if (error instanceof ApiError) {
      return
    }
    const message = error instanceof Error && error.message.trim() ? error.message : fallback
    toast.error(message)
  }

  async function downloadExportedReport(kind: ExportKind, reportId: string): Promise<void> {
    const successMessage =
      kind === 'archive' ? '复盘归档已生成并开始下载' : '教师复盘报告已生成并开始下载'
    const fallbackMessage =
      kind === 'archive' ? '复盘归档下载失败，请稍后重试' : '教师复盘报告下载失败，请稍后重试'

    try {
      await downloadGeneratedReport(reportId)
      toast.success(successMessage)
    } catch (error) {
      notifyExportError(error, fallbackMessage)
    }
  }

  function handleExportUpdate(kind: ExportKind, next: ReportExportData): void {
    if (next.report_id !== pendingReportId.value) return

    if (next.status === 'ready') {
      pendingReportId.value = null
      stopPolling()
      void downloadExportedReport(kind, next.report_id)
      return
    }

    if (next.status === 'failed') {
      pendingReportId.value = null
      stopPolling()
      toast.error(
        next.error_message || (kind === 'archive' ? '复盘归档生成失败' : '教师复盘报告生成失败')
      )
    }
  }

  async function startExport(kind: ExportKind): Promise<void> {
    if (!contestId.value) return
    if (kind === 'report' && !canExportReport.value) return

    exporting.value = kind
    try {
      const payload = buildExportPayload()
      const result =
        kind === 'archive'
          ? await exportTeacherAWDReviewArchive(contestId.value, payload)
          : await exportTeacherAWDReviewReport(contestId.value, payload)

      if (result.status === 'ready') {
        pendingReportId.value = null
        stopPolling()
        await downloadExportedReport(kind, result.report_id)
        return
      }

      if (result.status === 'failed') {
        stopPolling()
        toast.error(
          result.error_message || (kind === 'archive' ? '复盘归档生成失败' : '教师复盘报告生成失败')
        )
        return
      }

      pendingReportId.value = result.report_id
      startPolling(
        result.report_id,
        (next) => {
          handleExportUpdate(kind, next)
        },
        (error) => {
          pendingReportId.value = null
          notifyExportError(
            error,
            kind === 'archive'
              ? '复盘归档生成状态同步失败，请稍后重试'
              : '教师复盘报告生成状态同步失败，请稍后重试'
          )
        }
      )
      toast.info(
        kind === 'archive'
          ? '复盘归档开始生成，完成后会自动下载'
          : '教师复盘报告开始生成，完成后会自动下载'
      )
    } catch (error) {
      pendingReportId.value = null
      stopPolling()
      notifyExportError(
        error,
        kind === 'archive' ? '复盘归档导出失败，请稍后重试' : '教师复盘报告导出失败，请稍后重试'
      )
    } finally {
      exporting.value = null
    }
  }

  async function exportArchive(): Promise<void> {
    await startExport('archive')
  }

  async function exportReport(): Promise<void> {
    await startExport('report')
  }

  return {
    polling,
    exporting,
    pendingReportId,
    exportArchive,
    exportReport,
  }
}
