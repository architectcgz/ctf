import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { downloadReport } from '@/api/assessment'
import { ApiError } from '@/api/request'
import { exportStudentReviewArchive } from '@/api/teacher'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import {
  resolveClassStudentsRouteName,
  resolveStudentAnalysisRouteName,
  resolveStudentManagementRouteName,
} from '@/utils/teachingWorkspaceRouting'
import { useTeacherStudentReviewArchive } from './useTeacherStudentReviewArchive'

export function useTeacherStudentReviewArchivePage() {
  const route = useRoute()
  const router = useRouter()
  const toast = useToast()
  const authStore = useAuthStore()
  const { start: startPolling, stop: stopPolling } = useReportStatusPolling()

  const className = computed(() => String(route.params.className || ''))
  const studentId = computed(() => String(route.params.studentId || ''))
  const { archive, loading, error, reload } = useTeacherStudentReviewArchive(studentId)

  const exporting = ref(false)
  const pendingReportId = ref<string | null>(null)

  function openStudentAnalysis(): void {
    if (!studentId.value || !className.value) return
    router.push({
      name: resolveStudentAnalysisRouteName(authStore.user?.role),
      params: {
        className: className.value,
        studentId: studentId.value,
      },
    })
  }

  function goBack(): void {
    if (!className.value) {
      router.push({ name: resolveStudentManagementRouteName(authStore.user?.role) })
      return
    }
    router.push({
      name: resolveClassStudentsRouteName(authStore.user?.role),
      params: { className: className.value },
    })
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

  function notifyExportActionError(error: unknown, fallback: string): void {
    console.error(fallback, error)
    if (error instanceof ApiError) {
      return
    }
    const message = error instanceof Error && error.message.trim() ? error.message : fallback
    toast.error(message)
  }

  async function downloadArchiveReport(reportId: string): Promise<void> {
    try {
      await downloadGeneratedReport(reportId)
      toast.success('复盘归档已生成并开始下载')
    } catch (error) {
      notifyExportActionError(error, '复盘归档下载失败，请稍后重试')
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
        toast.error(result.error_message || '复盘归档生成失败')
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
            toast.error(next.error_message || '复盘归档生成失败')
          }
        },
        (error) => {
          pendingReportId.value = null
          notifyExportActionError(error, '复盘归档生成状态同步失败，请稍后重试')
        }
      )
      toast.info('复盘归档开始生成，完成后会自动下载')
    } catch (error) {
      pendingReportId.value = null
      stopPolling()
      notifyExportActionError(error, '复盘归档导出失败，请稍后重试')
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
    openStudentAnalysis,
    goBack,
    exportArchive,
  }
}
