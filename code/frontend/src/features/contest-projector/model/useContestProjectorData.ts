import { computed, ref } from 'vue'

import {
  getAdminContestLiveScoreboard,
  getContests,
  listContestAWDRounds,
} from '@/api/admin/contests'
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
import { useProjectorRoundSelection } from './useProjectorRoundSelection'
import { useProjectorRoundSnapshotLoader } from './useProjectorRoundSnapshotLoader'

export function useContestProjectorData() {
  const toast = useToast()
  let scoreboardRequestToken = 0

  const contests = ref<ContestDetailData[]>([])
  const selectedContestId = ref('')
  const scoreboard = ref<ContestScoreboardData | null>(null)
  const rounds = ref<AWDRoundData[]>([])
  const selectedRoundId = ref('')
  const roundAutoFollow = ref(true)
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
  const { chooseDisplayRound, enableAutoFollow } = useProjectorRoundSelection({
    roundAutoFollow,
    selectedRoundId,
  })
  const { loadRoundSnapshot, clearRoundSnapshot } = useProjectorRoundSnapshotLoader({
    services,
    attacks,
    roundSummary,
    trafficSummary,
  })

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
      const preferredRound = chooseDisplayRound(nextRounds)
      selectedRoundId.value = preferredRound?.id ?? ''

      if (preferredRound) {
        await loadRoundSnapshot(contestId, preferredRound.id, requestToken, (token) => token !== scoreboardRequestToken)
        if (requestToken !== scoreboardRequestToken) {
          return
        }
      } else {
        clearRoundSnapshot()
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
      roundAutoFollow.value = true
      clearRoundSnapshot()
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
      rounds.value = []
      selectedRoundId.value = ''
      roundAutoFollow.value = true
      clearRoundSnapshot()
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
    roundAutoFollow.value = true
    selectedRoundId.value = ''
    await loadScoreboard(contestId)
  }

  async function selectRound(roundId: string): Promise<void> {
    if (!selectedContestId.value || !roundId || (roundId === selectedRoundId.value && !roundAutoFollow.value)) {
      return
    }
    const targetRound = rounds.value.find((item) => item.id === roundId)
    if (!targetRound) {
      return
    }

    const requestToken = ++scoreboardRequestToken
    roundAutoFollow.value = false
    selectedRoundId.value = targetRound.id
    loadingScoreboard.value = true
    try {
      await loadRoundSnapshot(
        selectedContestId.value,
        targetRound.id,
        requestToken,
        (token) => token !== scoreboardRequestToken
      )
      if (requestToken !== scoreboardRequestToken) {
        return
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
      toast.error('同步大屏轮次失败')
    } finally {
      if (requestToken === scoreboardRequestToken) {
        loadingScoreboard.value = false
      }
    }
  }

  async function followCurrentRound(): Promise<void> {
    if (!enableAutoFollow()) {
      return
    }
    await loadScoreboard()
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
    roundAutoFollow,
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
    selectRound,
    followCurrentRound,
    startAutoRefresh,
    stopAutoRefresh,
  }
}
