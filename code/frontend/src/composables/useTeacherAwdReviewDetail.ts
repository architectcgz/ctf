import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { downloadReport } from '@/api/assessment'
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

  const loading = ref(false)
  const error = ref<string | null>(null)
  const review = ref<TeacherAWDReviewArchiveData | null>(null)
  const exporting = ref<ExportKind | null>(null)
  const pendingReportId = ref<string | null>(null)
  const selectedTeamId = ref<string | null>(null)

  const contestId = computed(() => String(route.params.contestId || ''))
  const selectedRoundNumber = computed(() => parseRoundQuery(route.query.round))
  const selectedRound = computed(() => review.value?.selected_round)
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

      if (
        selectedTeamId.value &&
        !next.selected_round?.teams.some((item) => item.team_id === selectedTeamId.value)
      ) {
        selectedTeamId.value = null
      }
    } catch (err) {
      console.error('加载 AWD 复盘详情失败:', err)
      review.value = null
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

  function handleExportUpdate(kind: ExportKind, next: ReportExportData): void {
    if (next.report_id !== pendingReportId.value) return

    if (next.status === 'ready') {
      pendingReportId.value = null
      stopPolling()
      void downloadGeneratedReport(next.report_id)
      toast.success(
        kind === 'archive' ? '复盘归档已生成并开始下载' : '教师复盘报告已生成并开始下载'
      )
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
        stopPolling()
        await downloadGeneratedReport(result.report_id)
        toast.success(
          kind === 'archive' ? '复盘归档已生成并开始下载' : '教师复盘报告已生成并开始下载'
        )
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
      })
      toast.info(
        kind === 'archive'
          ? '复盘归档开始生成，完成后会自动下载'
          : '教师复盘报告开始生成，完成后会自动下载'
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

  function openTeam(team: TeacherAWDReviewTeamItemData): void {
    selectedTeamId.value = team.team_id
  }

  function closeTeam(): void {
    selectedTeamId.value = null
  }

  watch(
    () => [route.params.contestId, route.query.round],
    () => {
      void loadReview()
    },
    { immediate: true }
  )

  return {
    router,
    route,
    polling,
    loading,
    error,
    review,
    exporting,
    contestId,
    selectedRoundNumber,
    selectedRound,
    selectedTeamId,
    selectedTeam,
    selectedTeamServices,
    selectedTeamAttacks,
    selectedTeamTraffic,
    canExportReport,
    loadReview,
    setRound,
    openTeam,
    closeTeam,
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
