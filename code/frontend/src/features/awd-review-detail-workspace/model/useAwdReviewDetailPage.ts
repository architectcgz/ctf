import { computed, onUnmounted, watch } from 'vue'

import { getReportStatus } from '@/api/assessment'
import { useRouteNavigationTransport } from '@/shared/model/navigation/useRouteNavigationTransport'
import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'
import { useBackofficeBreadcrumbDetail } from '@/shared/model/layout/useBackofficeBreadcrumbDetail'
import { useReportStatusPolling } from '@/shared/model/reporting/useReportStatusPolling'
import { useAuthStore } from '@/stores/auth'
import { useAwdReviewExportFlow } from '@/features/awd-review-workspace'
import { awdReviewIndexRoute } from './awdReviewDetailRoutes'
import { useAwdReviewDetailData } from './useAwdReviewDetailData'

export function useAwdReviewDetailPage() {
  const { params, query, replaceQuery } = useRouteQueryTransport()
  const { push } = useRouteNavigationTransport()
  const authStore = useAuthStore()
  const { polling, start: startPolling, stop: stopPolling } = useReportStatusPolling(getReportStatus)
  const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

  const contestId = computed(() => String(params.value.contestId || ''))
  const selectedRoundNumber = computed(() => parseRoundQuery(query.value.round))
  const data = useAwdReviewDetailData({
    contestId,
    selectedRoundNumber,
  })
  const { exporting, exportArchive, exportReport } = useAwdReviewExportFlow({
    contestId,
    selectedRoundNumber,
    canExportReport: data.canExportReport,
    startPolling,
    stopPolling,
  })

  function setRound(roundNumber?: number): void {
    const nextQuery = {
      ...query.value,
    } as Record<string, string>

    if (roundNumber) {
      nextQuery.round = String(roundNumber)
    } else {
      delete nextQuery.round
      delete nextQuery.team_id
    }

    void replaceQuery(nextQuery)
  }

  function openReviewIndex(): void {
    void push(awdReviewIndexRoute(authStore.user?.role))
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
    () => [params.value.contestId, query.value.round],
    () => {
      void data.loadReview()
    },
    { immediate: true }
  )

  watch(
    () => data.review.value?.contest.title,
    (title) => {
      if (title) {
        setBreadcrumbDetailTitle(title)
        return
      }
      setBreadcrumbDetailTitle()
    },
    { immediate: true }
  )

  onUnmounted(() => {
    setBreadcrumbDetailTitle()
  })

  return {
    polling,
    loading: data.loading,
    error: data.error,
    review: data.review,
    exporting,
    contestId,
    activeContestTitle: data.activeContestTitle,
    activeSummaryTitle: data.activeSummaryTitle,
    summaryStats: data.summaryStats,
    timelineRounds: data.timelineRounds,
    selectedRoundNumber,
    selectedRound: data.selectedRound,
    selectedTeamId: data.selectedTeamId,
    selectedTeam: data.selectedTeam,
    selectedTeamServices: data.selectedTeamServices,
    selectedTeamAttacks: data.selectedTeamAttacks,
    selectedTeamTraffic: data.selectedTeamTraffic,
    canExportReport: data.canExportReport,
    openReviewIndex,
    loadReview: data.loadReview,
    setRound,
    openTeam: data.openTeam,
    closeTeam: data.closeTeam,
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
