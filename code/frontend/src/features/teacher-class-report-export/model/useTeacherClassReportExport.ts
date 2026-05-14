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
import {
  buildTeacherClassInsightWindowQuery,
  createTeacherClassInsightWindowDraft,
  describeTeacherClassInsightWindow,
  getTeacherClassInsightWindowError,
} from '@/features/teacher-class-insight-window/model/window'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/format'

type ReportFormat = 'pdf' | 'excel'

interface ExportRecord {
  className: string
  format: ReportFormat
  fromDate: string
  toDate: string
  createdAt: string
  result: ReportExportData
}

interface ExportContext {
  className?: string
  fromDate?: string
  toDate?: string
}

export function useTeacherClassReportExport() {
  const authStore = useAuthStore()
  const toast = useToast()
  const { polling, start: startPolling, stop: stopPolling } = useReportStatusPolling()

  const form = ref({
    className: authStore.user?.class_name ?? '',
    format: 'pdf' as ReportFormat,
    fromDate: '',
    toDate: '',
  })

  const submitting = ref(false)
  const downloading = ref(false)
  const latestExport = ref<ExportRecord | null>(null)
  const previewLoading = ref(false)
  const previewError = ref<string | null>(null)
  const previewClassName = ref('')
  const previewStudents = ref<TeacherStudentItem[]>([])
  const previewReview = ref<TeacherClassReviewData | null>(null)
  const previewSummary = ref<TeacherClassSummaryData | null>(null)
  const previewTrend = ref<TeacherClassTrendData | null>(null)
  let latestPreviewRequestId = 0

  const classNamePlaceholder = computed(() =>
    authStore.user?.class_name ? `默认班级：${authStore.user.class_name}` : '请输入要导出的班级名称'
  )

  const normalizedClassNameText = computed(() => normalizeClassName() || '未选择')
  const selectedWindowLabel = computed(() =>
    describeTeacherClassInsightWindow(currentInsightWindow())
  )
  const selectedWindowError = computed(() =>
    getTeacherClassInsightWindowError(currentInsightWindow())
  )
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
        chipClass: 'class-report-task-chip--idle',
      }
    }

    switch (latestExport.value.result.status) {
      case 'ready':
        return {
          label: '已就绪',
          chipClass: 'class-report-task-chip--ready',
        }
      case 'failed':
        return {
          label: '失败',
          chipClass: 'class-report-task-chip--failed',
        }
      default:
        return {
          label: polling.value ? '生成中' : '等待更新',
          chipClass: 'class-report-task-chip--pending',
        }
    }
  })

  const latestExpiresText = computed(() => {
    if (!latestExport.value) return '--'
    return latestExport.value.result.expires_at
      ? formatDate(latestExport.value.result.expires_at)
      : '待生成完成后返回'
  })
  const latestWindowLabel = computed(() => {
    if (!latestExport.value) return '默认最近 7 天'
    return describeTeacherClassInsightWindow({
      fromDate: latestExport.value.fromDate,
      toDate: latestExport.value.toDate,
    })
  })

  function resolveContextClassName(className?: string): string {
    return className?.trim() || authStore.user?.class_name?.trim() || ''
  }

  function normalizeClassName(): string {
    return form.value.className.trim() || authStore.user?.class_name?.trim() || ''
  }

  function currentInsightWindow() {
    return createTeacherClassInsightWindowDraft({
      fromDate: form.value.fromDate,
      toDate: form.value.toDate,
    })
  }

  function resetPreviewState(): void {
    previewStudents.value = []
    previewReview.value = null
    previewSummary.value = null
    previewTrend.value = null
  }

  function syncContext(context?: ExportContext): void {
    const nextClassName = resolveContextClassName(context?.className)
    const nextInsightWindow = createTeacherClassInsightWindowDraft({
      fromDate: context?.fromDate,
      toDate: context?.toDate,
    })
    form.value.className = nextClassName
    form.value.fromDate = nextInsightWindow.fromDate
    form.value.toDate = nextInsightWindow.toDate

    if (latestExport.value && latestExport.value.className !== nextClassName) {
      stopPolling()
      latestExport.value = null
    }
  }

  async function loadPreview(): Promise<void> {
    const className = normalizeClassName()
    if (!className) {
      previewClassName.value = ''
      resetPreviewState()
      previewError.value = '请先填写班级名称'
      return
    }

    const insightWindow = currentInsightWindow()
    const insightWindowError = getTeacherClassInsightWindowError(insightWindow)
    if (insightWindowError) {
      previewClassName.value = className
      previewError.value = insightWindowError
      return
    }

    const requestId = ++latestPreviewRequestId
    const insightWindowQuery = buildTeacherClassInsightWindowQuery(insightWindow)
    previewLoading.value = true
    previewError.value = null
    previewClassName.value = className

    try {
      const [students, review, summary, trend] = await Promise.all([
        getClassStudents(className),
        insightWindowQuery
          ? getClassReview(className, insightWindowQuery)
          : getClassReview(className),
        insightWindowQuery
          ? getClassSummary(className, insightWindowQuery)
          : getClassSummary(className),
        insightWindowQuery
          ? getClassTrend(className, insightWindowQuery)
          : getClassTrend(className),
      ])
      if (requestId !== latestPreviewRequestId) {
        return
      }
      previewStudents.value = students
      previewReview.value = review
      previewSummary.value = summary
      previewTrend.value = trend
    } catch (err) {
      if (requestId !== latestPreviewRequestId) {
        return
      }
      console.error('加载班级报告预览失败:', err)
      resetPreviewState()
      previewError.value = '加载当前班级预览失败，请稍后重试'
    } finally {
      if (requestId === latestPreviewRequestId) {
        previewLoading.value = false
      }
    }
  }

  async function handleExport(): Promise<void> {
    if (submitting.value) {
      return
    }

    const className = normalizeClassName()
    if (!className) {
      toast.warning('请先填写班级名称')
      return
    }

    const insightWindow = currentInsightWindow()
    const insightWindowError = getTeacherClassInsightWindowError(insightWindow)
    if (insightWindowError) {
      toast.warning(insightWindowError)
      return
    }

    submitting.value = true
    try {
      const insightWindowQuery = buildTeacherClassInsightWindowQuery(insightWindow)
      const result = await exportClassReport({
        class_name: className,
        format: form.value.format,
        ...insightWindowQuery,
      })

      latestExport.value = {
        className,
        format: form.value.format,
        fromDate: insightWindow.fromDate,
        toDate: insightWindow.toDate,
        createdAt: new Date().toISOString(),
        result,
      }

      if (result.status === 'ready') {
        stopPolling()
        toast.success('报告已生成，可立即下载')
        return
      }

      if (result.status === 'failed') {
        stopPolling()
        toast.error(result.error_message || '报告生成失败')
        return
      }

      startPolling(String(result.report_id), (next) => {
        if (!latestExport.value) return
        latestExport.value = {
          ...latestExport.value,
          result: next,
        }
      })
      toast.info('报告开始生成，系统会自动刷新任务状态')
    } catch (err) {
      console.error('创建班级报告导出任务失败:', err)
      toast.error('创建导出任务失败，请稍后重试')
    } finally {
      submitting.value = false
    }
  }

  async function handleDownload(): Promise<void> {
    if (downloading.value || !latestExport.value || latestExport.value.result.status !== 'ready') {
      return
    }

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
    } catch (err) {
      console.error('下载班级报告失败:', err)
      toast.error('下载报告失败，请稍后重试')
    } finally {
      downloading.value = false
    }
  }

  return {
    polling,
    form,
    submitting,
    downloading,
    latestExport,
    previewLoading,
    previewError,
    previewClassName,
    previewStudents,
    previewReview,
    previewSummary,
    previewTrend,
    classNamePlaceholder,
    normalizedClassNameText,
    selectedWindowLabel,
    selectedWindowError,
    selectedFormatLabel,
    selectedFormatHint,
    derivedDownloadHint,
    averageSolvedText,
    activeRateText,
    latestStatusMeta,
    latestExpiresText,
    latestWindowLabel,
    syncContext,
    loadPreview,
    handleExport,
    handleDownload,
  }
}
