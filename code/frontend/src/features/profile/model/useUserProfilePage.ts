import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Activity, ShieldCheck, UserCircle2 } from 'lucide-vue-next'

import { downloadReport, exportPersonalReport } from '@/api/assessment'
import { getProfile } from '@/api/auth'
import type { AuthUser, ReportExportData } from '@/api/contracts'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/format'

export function useUserProfilePage() {
  const authStore = useAuthStore()

  const loading = ref(false)
  const error = ref<string | null>(null)
  const profile = ref<AuthUser | null>(null)
  const exportLoading = ref(false)
  const exportError = ref<string | null>(null)
  const reportFormat = ref<'pdf' | 'excel'>('pdf')
  const latestReport = ref<ReportExportData | null>(null)
  const latestReportFormat = ref<'pdf' | 'excel'>('pdf')
  const latestReportCreatedAt = ref<string | null>(null)
  const { start: startPolling, stop: stopPolling } = useReportStatusPolling()
  const currentRole = computed(() => profile.value?.role ?? authStore.user?.role)
  const canManagePersonalReport = computed(() => currentRole.value !== 'admin')
  const currentProfile = computed(() => profile.value ?? authStore.user ?? null)
  const pageCopy = computed(() =>
    canManagePersonalReport.value
      ? '查看账号信息、个人报告与最近导出状态。'
      : '查看账号信息与当前账号状态。'
  )

  function getRoleLabel(role: AuthUser['role'] | undefined): string {
    if (role === 'admin') return '管理员'
    if (role === 'teacher') return '教师'
    if (role === 'student') return '学生'
    return '未知'
  }

  const profileFields = computed(() => {
    const current = currentProfile.value
    if (!current) return []
    return [
      { label: 'Username', value: current.username },
      { label: 'Role', value: current.role },
      { label: 'Class', value: current.class_name || '未分配' },
      { label: 'Name', value: current.name || '未填写' },
    ]
  })

  const reportTaskMeta = computed(() => {
    if (latestReport.value?.status === 'ready') {
      return { label: '可下载', status: 'ready', chipClass: 'chip--success' }
    }
    if (latestReport.value?.status === 'failed') {
      return { label: '生成失败', status: 'failed', chipClass: 'chip--danger' }
    }
    if (latestReport.value?.status === 'processing') {
      return { label: '生成中', status: 'processing', chipClass: 'chip--warning' }
    }
    return { label: '待创建', status: 'idle', chipClass: 'chip--primary' }
  })

  const profileSummaryItems = computed(() => {
    const current = currentProfile.value
    const className = current?.class_name || '未分配'
    const displayName = current?.name || '未填写'

    return [
      {
        key: 'status',
        label: '账号状态',
        value: '正常',
        helper: '当前账号可正常访问个人工作区',
        icon: ShieldCheck,
        techFont: false,
      },
      {
        key: 'role',
        label: '当前角色',
        value: getRoleLabel(current?.role),
        helper: '决定当前账号可访问的功能范围',
        icon: UserCircle2,
        techFont: false,
      },
      {
        key: canManagePersonalReport.value ? 'report' : 'name',
        label: canManagePersonalReport.value ? '报告状态' : '实名信息',
        value: canManagePersonalReport.value ? reportTaskMeta.value.label : displayName,
        helper: canManagePersonalReport.value
          ? latestReportCreatedAt.value
            ? `最近生成于 ${formatDate(latestReportCreatedAt.value)}`
            : '当前还没有生成过个人报告'
          : current?.name
            ? '用于账号展示与身份识别'
            : '当前未填写姓名信息',
        icon: canManagePersonalReport.value ? Activity : UserCircle2,
        techFont: false,
      },
      {
        key: 'class',
        label: '所属班级',
        value: className,
        helper: current?.class_name ? '当前归属的班级信息' : '当前账号还未绑定班级',
        icon: UserCircle2,
        techFont: false,
      },
    ]
  })

  async function loadProfile(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      profile.value = await getProfile()
    } catch (err) {
      console.error('加载个人资料失败:', err)
      profile.value = authStore.user ? { ...authStore.user } : null
      error.value = '加载个人资料失败，以下展示的是本地缓存信息'
    } finally {
      loading.value = false
    }
  }

  async function createReport(): Promise<void> {
    exportLoading.value = true
    exportError.value = null
    try {
      latestReportFormat.value = reportFormat.value
      latestReportCreatedAt.value = new Date().toISOString()
      latestReport.value = await exportPersonalReport({ format: reportFormat.value })
      if (latestReport.value.status === 'processing') {
        startPolling(String(latestReport.value.report_id), (next) => {
          latestReport.value = next
        })
      } else {
        stopPolling()
        if (latestReport.value.status === 'failed') {
          exportError.value = latestReport.value.error_message || '个人报告生成失败，请稍后重试'
        }
      }
    } catch (err) {
      console.error('导出个人报告失败:', err)
      exportError.value = '创建个人报告失败，请稍后重试'
    } finally {
      exportLoading.value = false
    }
  }

  async function handleDownload(): Promise<void> {
    if (!latestReport.value) return
    exportError.value = null
    try {
      const file = await downloadReport(latestReport.value.report_id)
      const url = window.URL.createObjectURL(file.blob)
      const link = document.createElement('a')
      link.href = url
      link.download = file.filename
      link.click()
      window.URL.revokeObjectURL(url)
    } catch (err) {
      console.error('下载个人报告失败:', err)
      exportError.value =
        err instanceof Error && err.message.trim() ? err.message : '下载最近报告失败，请稍后重试'
    }
  }

  onMounted(() => {
    void loadProfile()
  })

  onUnmounted(() => {
    stopPolling()
  })

  return {
    loading,
    error,
    profile,
    exportLoading,
    exportError,
    reportFormat,
    latestReport,
    latestReportFormat,
    latestReportCreatedAt,
    canManagePersonalReport,
    currentProfile,
    pageCopy,
    profileFields,
    reportTaskMeta,
    profileSummaryItems,
    loadProfile,
    createReport,
    handleDownload,
  }
}
