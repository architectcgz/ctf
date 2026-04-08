import { computed, ref } from 'vue'

import { downloadReport } from '@/api/assessment'
import {
  exportClassReport,
  getClassReview,
  getClassStudents,
  getClassSummary,
  getClassTrend,
} from '@/api/teacher'
import type {
  ReportExportData,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/format'

type ReportFormat = 'pdf' | 'excel'

interface ExportRecord {
  className: string
  format: ReportFormat
  createdAt: string
  result: ReportExportData
}

export function useTeacherReportExportPage() {
  const authStore = useAuthStore()
  const toast = useToast()
  const { polling, start: startPolling, stop: stopPolling } = useReportStatusPolling()

  const form = ref({
    className: authStore.user?.class_name ?? '',
    format: 'pdf' as ReportFormat,
  })

  const submitting = ref(false)
  const downloading = ref(false)
  const latestExport = ref<ExportRecord | null>(null)
  const previewDialogVisible = ref(false)
  const previewLoading = ref(false)
  const previewError = ref<string | null>(null)
  const previewClassName = ref('')
  const previewStudents = ref<TeacherStudentItem[]>([])
  const previewReview = ref<TeacherClassReviewData | null>(null)
  const previewSummary = ref<TeacherClassSummaryData | null>(null)
  const previewTrend = ref<TeacherClassTrendData | null>(null)

  const classNamePlaceholder = computed(() =>
    authStore.user?.class_name ? `默认班级：${authStore.user.class_name}` : '请输入要导出的班级名称'
  )

  const normalizedClassNameText = computed(() => normalizeClassName() || '未选择')
  const selectedFormatLabel = computed(() => (form.value.format === 'pdf' ? 'PDF' : 'Excel'))

  const selectedFormatHint = computed(() =>
    form.value.format === 'pdf' ? '适合打印、归档和正式汇报。' : '适合继续分析、筛选和二次加工。'
  )

  const derivedDownloadHint = computed(() => {
    if (!latestExport.value) return ''
    if (latestExport.value.result.status === 'ready') {
      return '报告已生成，可直接下载。'
    }
    if (latestExport.value.result.status === 'failed') {
      return latestExport.value.result.error_message || '报告生成失败，请重新发起导出任务。'
    }
    return '正在轮询导出状态，生成完成后会自动更新为可下载。'
  })

  const averageSolvedText = computed(() => {
    if (!previewSummary.value) return '--'
    return previewSummary.value.average_solved.toFixed(1)
  })

  const activeRateText = computed(() => {
    if (!previewSummary.value) return '--'
    return `${Math.round(previewSummary.value.active_rate)}%`
  })

  const latestStatusMeta = computed(() => {
    if (!latestExport.value) {
      return {
        label: '未创建',
        chipClass: 'report-status-chip--idle',
      }
    }

    switch (latestExport.value.result.status) {
      case 'ready':
        return {
          label: '已就绪',
          chipClass: 'report-status-chip--ready',
        }
      case 'failed':
        return {
          label: '失败',
          chipClass: 'report-status-chip--failed',
        }
      default:
        return {
          label: polling.value ? '生成中' : '等待更新',
          chipClass: 'report-status-chip--pending',
        }
    }
  })

  const latestExpiresText = computed(() => {
    if (!latestExport.value) return '--'
    return latestExport.value.result.expires_at
      ? formatDate(latestExport.value.result.expires_at)
      : '待生成完成后返回'
  })

  function normalizeClassName(): string {
    return form.value.className.trim() || authStore.user?.class_name?.trim() || ''
  }

  async function loadPreview(): Promise<void> {
    const className = normalizeClassName()
    if (!className) {
      previewClassName.value = ''
      previewStudents.value = []
      previewReview.value = null
      previewSummary.value = null
      previewTrend.value = null
      previewError.value = '请先填写班级名称'
      return
    }

    previewLoading.value = true
    previewError.value = null
    previewClassName.value = className

    try {
      const [students, review, summary, trend] = await Promise.all([
        getClassStudents(className),
        getClassReview(className),
        getClassSummary(className),
        getClassTrend(className),
      ])
      previewStudents.value = students
      previewReview.value = review
      previewSummary.value = summary
      previewTrend.value = trend
    } catch (err) {
      console.error('加载报告预览失败:', err)
      previewStudents.value = []
      previewReview.value = null
      previewSummary.value = null
      previewTrend.value = null
      previewError.value = '加载当前班级预览失败，请稍后重试'
    } finally {
      previewLoading.value = false
    }
  }

  async function openPreviewDialog(): Promise<void> {
    previewDialogVisible.value = true
    await loadPreview()
  }

  async function handleExport(): Promise<void> {
    const className = normalizeClassName()
    if (!className) {
      toast.warning('请先填写班级名称')
      return
    }

    submitting.value = true
    try {
      const result = await exportClassReport({
        class_name: className,
        format: form.value.format,
      })

      latestExport.value = {
        className,
        format: form.value.format,
        createdAt: new Date().toISOString(),
        result,
      }

      if (result.status === 'ready') {
        stopPolling()
        toast.success('报告已生成，可立即下载')
      } else if (result.status === 'failed') {
        stopPolling()
        toast.error(result.error_message || '报告生成失败')
      } else {
        startPolling(String(result.report_id), (next) => {
          if (!latestExport.value) return
          latestExport.value = {
            ...latestExport.value,
            result: next,
          }
        })
        toast.info('报告开始生成，系统会自动刷新任务状态')
      }
    } finally {
      submitting.value = false
    }
  }

  async function handleDownload(): Promise<void> {
    if (!latestExport.value) return

    downloading.value = true
    try {
      const { blob, filename } = await downloadReport(latestExport.value.result.report_id)
      const objectUrl = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = objectUrl
      link.download = filename
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(objectUrl)
      toast.success('下载已开始')
    } finally {
      downloading.value = false
    }
  }

  return {
    authStore,
    polling,
    form,
    submitting,
    downloading,
    latestExport,
    previewDialogVisible,
    previewLoading,
    previewError,
    previewClassName,
    previewStudents,
    previewReview,
    previewSummary,
    previewTrend,
    classNamePlaceholder,
    normalizedClassNameText,
    selectedFormatLabel,
    selectedFormatHint,
    derivedDownloadHint,
    averageSolvedText,
    activeRateText,
    latestStatusMeta,
    latestExpiresText,
    openPreviewDialog,
    handleExport,
    handleDownload,
  }
}
