import { computed, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { downloadReport } from '@/api/assessment'
import { ApiError } from '@/api/request'
import {
  exportTeacherAWDReviewArchive,
  exportTeacherAWDReviewReport,
  getTeacherAWDReview,
} from '@/api/teacher'
import type {
  ReportExportData,
  TeacherAWDReviewArchiveData,
  TeacherAWDReviewTeamItemData,
} from '@/api/contracts'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import { resolveAwdReviewDetailRouteName } from '@/utils/teachingWorkspaceRouting'

type ExportKind = 'archive' | 'report'

export function useTeacherAwdReviewDetail() {
  const route = useRoute()
  const router = useRouter()
  const toast = useToast()
  const authStore = useAuthStore()
  const { polling, start: startPolling, stop: stopPolling } = useReportStatusPolling()
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const loading = ref(false)
  const error = ref<string | null>(null)
  const review = ref<TeacherAWDReviewArchiveData | null>(null)
  const exporting = ref<ExportKind | null>(null)
  const pendingReportId = ref<string | null>(null)
  const selectedTeamId = ref<string | null>(null)

  const contestId = computed(() => String(route.params.contestId || ''))
  const selectedRoundNumber = computed(() => parseRoundQuery(route.query.round))
  const selectedRound = computed(() => review.value?.selected_round)
  const activeContestTitle = computed(() => review.value?.contest.title || '--')
  const activeSummaryTitle = computed(() =>
    selectedRoundNumber.value ? `第 ${selectedRoundNumber.value} 轮` : '整场总览'
  )
  const summaryStats = computed(() => {
    if (selectedRound.value) {
      return {
        roundCount: 1,
        teamCount: selectedRound.value.teams.length,
        serviceCount: selectedRound.value.round.service_count,
        attackCount: selectedRound.value.round.attack_count,
        trafficCount: selectedRound.value.round.traffic_count,
      }
    }

    return {
      roundCount: review.value?.overview?.round_count ?? 0,
      teamCount: review.value?.overview?.team_count ?? 0,
      serviceCount: review.value?.overview?.service_count ?? 0,
      attackCount: review.value?.overview?.attack_count ?? 0,
      trafficCount: review.value?.overview?.traffic_count ?? 0,
    }
  })
  const timelineRounds = computed(() => review.value?.rounds || [])
  const selectedTeam = computed(
    () => selectedRound.value?.teams.find((item) => item.team_id === selectedTeamId.value) ?? null
  )
  const selectedTeamServices = computed(
    () =>
      selectedRound.value?.services.filter((item) => item.team_id === selectedTeamId.value) ?? []
  )
  const selectedTeamAttacks = computed(
    () =>
      selectedRound.value?.attacks.filter(
        (item) =>
          item.attacker_team_id === selectedTeamId.value ||
          item.victim_team_id === selectedTeamId.value
      ) ?? []
  )
  const selectedTeamTraffic = computed(
    () =>
      selectedRound.value?.traffic.filter(
        (item) =>
          item.attacker_team_id === selectedTeamId.value ||
          item.victim_team_id === selectedTeamId.value
      ) ?? []
  )
  const canExportReport = computed(() => Boolean(review.value?.contest.export_ready))

  async function loadReview(): Promise<void> {
    if (!contestId.value) {
      review.value = null
      setBreadcrumbDetailTitle()
      return
    }

    loading.value = true
    error.value = null

    try {
      const next = await getTeacherAWDReview(contestId.value, {
        round: selectedRoundNumber.value,
        team_id: undefined,
      })
      review.value = next
      setBreadcrumbDetailTitle(next.contest.title)

      if (
        selectedTeamId.value &&
        !next.selected_round?.teams.some((item) => item.team_id === selectedTeamId.value)
      ) {
        selectedTeamId.value = null
      }
    } catch (err) {
      console.error('加载 AWD 复盘详情失败:', err)
      review.value = null
      setBreadcrumbDetailTitle()
      error.value = '加载 AWD 复盘详情失败，请稍后重试'
    } finally {
      loading.value = false
    }
  }

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
      startPolling(result.report_id, (next) => {
        handleExportUpdate(kind, next)
      }, (error) => {
        pendingReportId.value = null
        notifyExportError(
          error,
          kind === 'archive'
            ? '复盘归档生成状态同步失败，请稍后重试'
            : '教师复盘报告生成状态同步失败，请稍后重试'
        )
      })
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

  function setRound(roundNumber?: number): void {
    const query = {
      ...route.query,
    } as Record<string, string>

    if (roundNumber) {
      query.round = String(roundNumber)
    } else {
      delete query.round
      delete query.team_id
    }

    router.replace({
      name: resolveAwdReviewDetailRouteName(authStore.user?.role),
      params: {
        contestId: contestId.value,
      },
      query,
    })
  }

  function openReviewIndex(): void {
    router.push({ name: 'TeacherAWDReviewIndex' })
  }

  function openTeam(team: TeacherAWDReviewTeamItemData): void {
    selectedTeamId.value = team.team_id
  }

  function closeTeam(): void {
    selectedTeamId.value = null
  }

  function contestStatusLabel(status: string): string {
    switch (status) {
      case 'running':
        return '进行中'
      case 'ended':
        return '已结束'
      case 'frozen':
        return '冻结中'
      default:
        return status || '未开始'
    }
  }

  function formatServiceRef(serviceId?: string): string {
    return `Service #${serviceId || '--'}`
  }

  watch(
    () => [route.params.contestId, route.query.round],
    () => {
      void loadReview()
    },
    { immediate: true }
  )

  onUnmounted(() => {
    setBreadcrumbDetailTitle()
  })

  return {
    route,
    polling,
    loading,
    error,
    review,
    exporting,
    contestId,
    activeContestTitle,
    activeSummaryTitle,
    summaryStats,
    timelineRounds,
    selectedRoundNumber,
    selectedRound,
    selectedTeamId,
    selectedTeam,
    selectedTeamServices,
    selectedTeamAttacks,
    selectedTeamTraffic,
    canExportReport,
    openReviewIndex,
    loadReview,
    setRound,
    openTeam,
    closeTeam,
    contestStatusLabel,
    formatServiceRef,
    exportArchive,
    exportReport,
  }
}

function parseRoundQuery(value: unknown): number | undefined {
  const raw = Array.isArray(value) ? value[0] : value
  if (raw == null || raw === '') return undefined

  const normalized = Number(raw)
  if (!Number.isInteger(normalized) || normalized < 1) return undefined

  return normalized
}
