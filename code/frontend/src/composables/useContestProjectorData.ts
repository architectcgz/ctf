import { computed, ref } from 'vue'

import {
  getAdminContestLiveScoreboard,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  getContests,
  listContestAWDRoundAttacks,
  listContestAWDRoundServices,
  listContestAWDRounds,
} from '@/api/admin'
import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  ContestDetailData,
  ContestScoreboardData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

export function useContestProjectorData() {
  const toast = useToast()
  let scoreboardRequestToken = 0

  const contests = ref<ContestDetailData[]>([])
  const selectedContestId = ref('')
  const scoreboard = ref<ContestScoreboardData | null>(null)
  const rounds = ref<AWDRoundData[]>([])
  const selectedRoundId = ref('')
  const services = ref<AWDTeamServiceData[]>([])
  const attacks = ref<AWDAttackLogData[]>([])
  const roundSummary = ref<AWDRoundSummaryData | null>(null)
  const trafficSummary = ref<AWDTrafficSummaryData | null>(null)
  const loadingContests = ref(true)
  const loadingScoreboard = ref(false)
  const loadError = ref('')
  const refreshTimer = ref<number | null>(null)
  const lastUpdatedLabel = ref('未同步')

  const awdContests = computed(() => contests.value.filter((item) => item.mode === 'awd'))
  const projectorContests = computed(() =>
    awdContests.value
      .filter((item) => ['running', 'frozen', 'ended'].includes(item.status))
      .slice()
      .sort((left, right) => {
        const rightTime = new Date(right.starts_at ?? right.ends_at).getTime()
        const leftTime = new Date(left.starts_at ?? left.ends_at).getTime()
        return rightTime - leftTime
      })
  )
  const selectedContest = computed(
    () => projectorContests.value.find((item) => item.id === selectedContestId.value) ?? null
  )
  const scoreboardRows = computed(() => scoreboard.value?.scoreboard.list ?? [])
  const selectedRound = computed(() => rounds.value.find((item) => item.id === selectedRoundId.value) ?? null)

  function chooseInitialContest(): void {
    const preferred =
      projectorContests.value.find((item) => item.status === 'running') ??
      projectorContests.value.find((item) => item.status === 'frozen') ??
      projectorContests.value[0] ??
      null
    selectedContestId.value = preferred?.id ?? ''
  }

  async function loadScoreboard(contestId = selectedContestId.value): Promise<void> {
    if (!contestId) {
      return
    }

    const requestToken = ++scoreboardRequestToken
    loadingScoreboard.value = true
    try {
      const [nextScoreboard, nextRounds] = await Promise.all([
        getAdminContestLiveScoreboard(contestId, {
          page: 1,
          page_size: 20,
        }),
        listContestAWDRounds(contestId),
      ])
      if (requestToken !== scoreboardRequestToken) {
        return
      }
      scoreboard.value = nextScoreboard
      rounds.value = nextRounds
      const preferredRound =
        nextRounds.find((item) => item.status === 'running') ??
        nextRounds[nextRounds.length - 1] ??
        null
      selectedRoundId.value = preferredRound?.id ?? ''

      if (preferredRound) {
        const [nextServices, nextAttacks, nextRoundSummary, nextTrafficSummary] = await Promise.all([
          listContestAWDRoundServices(contestId, preferredRound.id),
          listContestAWDRoundAttacks(contestId, preferredRound.id),
          getContestAWDRoundSummary(contestId, preferredRound.id),
          getContestAWDRoundTrafficSummary(contestId, preferredRound.id),
        ])
        if (requestToken !== scoreboardRequestToken) {
          return
        }
        services.value = nextServices
        attacks.value = nextAttacks
        roundSummary.value = nextRoundSummary
        trafficSummary.value = nextTrafficSummary
      } else {
        services.value = []
        attacks.value = []
        roundSummary.value = null
        trafficSummary.value = null
      }
      lastUpdatedLabel.value = new Date().toLocaleTimeString('zh-CN', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
    } catch {
      if (requestToken !== scoreboardRequestToken) {
        return
      }
      scoreboard.value = null
      rounds.value = []
      selectedRoundId.value = ''
      services.value = []
      attacks.value = []
      roundSummary.value = null
      trafficSummary.value = null
      toast.error('同步大屏排行榜失败')
    } finally {
      if (requestToken === scoreboardRequestToken) {
        loadingScoreboard.value = false
      }
    }
  }

  async function loadContests(): Promise<void> {
    loadingContests.value = true
    loadError.value = ''
    try {
      const response = await getContests({
        page: 1,
        page_size: 100,
      })
      contests.value = response.list
      if (!selectedContestId.value || !projectorContests.value.some((item) => item.id === selectedContestId.value)) {
        chooseInitialContest()
      }
      await loadScoreboard()
    } catch (error) {
      contests.value = []
      scoreboard.value = null
      loadError.value = error instanceof Error ? error.message : '大屏赛事加载失败'
    } finally {
      loadingContests.value = false
    }
  }

  async function selectContest(contestId: string): Promise<void> {
    if (contestId === selectedContestId.value) {
      return
    }
    selectedContestId.value = contestId
    await loadScoreboard(contestId)
  }

  function startAutoRefresh(): void {
    if (refreshTimer.value !== null) {
      window.clearInterval(refreshTimer.value)
    }
    refreshTimer.value = window.setInterval(() => {
      void loadScoreboard()
    }, 15000)
  }

  function stopAutoRefresh(): void {
    if (refreshTimer.value !== null) {
      window.clearInterval(refreshTimer.value)
      refreshTimer.value = null
    }
  }

  return {
    contests,
    selectedContestId,
    scoreboard,
    rounds,
    selectedRoundId,
    services,
    attacks,
    roundSummary,
    trafficSummary,
    loadingContests,
    loadingScoreboard,
    loadError,
    lastUpdatedLabel,
    projectorContests,
    selectedContest,
    scoreboardRows,
    selectedRound,
    loadScoreboard,
    loadContests,
    selectContest,
    startAutoRefresh,
    stopAutoRefresh,
  }
}
