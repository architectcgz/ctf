import { computed, ref, type Ref } from 'vue'

import { getAwdReviewByRole } from '@/api/awd-reviews'
import type {
  AwdReviewArchiveData,
  AwdReviewTeamItemData,
} from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { reportFrontendError } from '@/utils/reportFrontendError'

interface UseAwdReviewDetailDataOptions {
  contestId: Readonly<Ref<string>>
  selectedRoundNumber: Readonly<Ref<number | undefined>>
}

export function useAwdReviewDetailData(options: UseAwdReviewDetailDataOptions) {
  const { contestId, selectedRoundNumber } = options
  const authStore = useAuthStore()

  const loading = ref(false)
  const error = ref<string | null>(null)
  const review = ref<AwdReviewArchiveData | null>(null)
  const selectedTeamId = ref<string | null>(null)

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
    () => selectedRound.value?.services.filter((item) => item.team_id === selectedTeamId.value) ?? []
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
      error.value = null
      review.value = null
      selectedTeamId.value = null
      return
    }

    loading.value = true
    error.value = null

    try {
      const next = await getAwdReviewByRole(authStore.user?.role, contestId.value, {
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
      reportFrontendError('加载 AWD 复盘详情失败:', err)
      review.value = null
      error.value = '加载 AWD 复盘详情失败，请稍后重试'
    } finally {
      loading.value = false
    }
  }

  function openTeam(team: AwdReviewTeamItemData): void {
    selectedTeamId.value = team.team_id
  }

  function closeTeam(): void {
    selectedTeamId.value = null
  }

  return {
    loading,
    error,
    review,
    selectedRound,
    activeContestTitle,
    activeSummaryTitle,
    summaryStats,
    timelineRounds,
    selectedTeamId,
    selectedTeam,
    selectedTeamServices,
    selectedTeamAttacks,
    selectedTeamTraffic,
    canExportReport,
    loadReview,
    openTeam,
    closeTeam,
  }
}
