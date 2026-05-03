import { ref, type Ref } from 'vue'

import {
  type AdminAWDTrafficEventsParams,
  getAdminContestLiveScoreboard,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  listContestAWDRoundAttacks,
  listContestAWDRoundServices,
  listContestAWDRoundTrafficEvents,
} from '@/api/admin/contests'
import type {
  AWDAttackLogData,
  AWDRoundSummaryData,
  AWDTrafficEventData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'

interface UseAwdRoundDetailStateOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
  selectedRoundId: Readonly<Ref<string | null>>
  buildTrafficEventsParams: () => AdminAWDTrafficEventsParams
  syncTrafficPagination: (page: number, pageSize: number) => void
}

export function useAwdRoundDetailState(options: UseAwdRoundDetailStateOptions) {
  const { selectedContest, selectedRoundId, buildTrafficEventsParams, syncTrafficPagination } =
    options

  const services = ref<AWDTeamServiceData[]>([])
  const attacks = ref<AWDAttackLogData[]>([])
  const summary = ref<AWDRoundSummaryData | null>(null)
  const trafficSummary = ref<AWDTrafficSummaryData | null>(null)
  const trafficEvents = ref<AWDTrafficEventData[]>([])
  const trafficEventsTotal = ref(0)
  const scoreboardRows = ref<ScoreboardRow[]>([])
  const scoreboardFrozen = ref(false)
  const loadingRoundDetail = ref(false)
  const loadingTrafficSummary = ref(false)
  const loadingTrafficEvents = ref(false)

  let roundDetailRequestToken = 0
  let trafficEventsRequestToken = 0

  function clearRoundDetail() {
    services.value = []
    attacks.value = []
    summary.value = null
    trafficSummary.value = null
    trafficEvents.value = []
    trafficEventsTotal.value = 0
    scoreboardRows.value = []
    scoreboardFrozen.value = false
  }

  async function refreshRoundDetail(roundId = selectedRoundId.value) {
    if (!selectedContest.value || !roundId) {
      clearRoundDetail()
      return
    }

    const requestToken = ++roundDetailRequestToken
    loadingRoundDetail.value = true
    loadingTrafficSummary.value = true
    loadingTrafficEvents.value = true
    try {
      const [
        nextServices,
        nextAttacks,
        nextSummary,
        nextTrafficSummary,
        nextTrafficEvents,
        nextScoreboard,
      ] = await Promise.all([
        listContestAWDRoundServices(selectedContest.value.id, roundId),
        listContestAWDRoundAttacks(selectedContest.value.id, roundId),
        getContestAWDRoundSummary(selectedContest.value.id, roundId),
        getContestAWDRoundTrafficSummary(selectedContest.value.id, roundId),
        listContestAWDRoundTrafficEvents(
          selectedContest.value.id,
          roundId,
          buildTrafficEventsParams()
        ),
        getAdminContestLiveScoreboard(selectedContest.value.id, { page: 1, page_size: 10 }),
      ])

      if (requestToken !== roundDetailRequestToken) {
        return
      }

      services.value = nextServices
      attacks.value = nextAttacks
      summary.value = nextSummary
      trafficSummary.value = nextTrafficSummary
      trafficEvents.value = nextTrafficEvents.list
      trafficEventsTotal.value = nextTrafficEvents.total
      syncTrafficPagination(nextTrafficEvents.page, nextTrafficEvents.page_size)
      scoreboardRows.value = nextScoreboard.scoreboard.list
      scoreboardFrozen.value = nextScoreboard.frozen
    } finally {
      if (requestToken === roundDetailRequestToken) {
        loadingRoundDetail.value = false
        loadingTrafficSummary.value = false
        loadingTrafficEvents.value = false
      }
    }
  }

  async function refreshTrafficEvents(roundId = selectedRoundId.value) {
    if (!selectedContest.value || !roundId) {
      trafficEvents.value = []
      trafficEventsTotal.value = 0
      return
    }

    const requestToken = ++trafficEventsRequestToken
    loadingTrafficEvents.value = true
    try {
      const result = await listContestAWDRoundTrafficEvents(
        selectedContest.value.id,
        roundId,
        buildTrafficEventsParams()
      )
      if (requestToken !== trafficEventsRequestToken) {
        return
      }
      trafficEvents.value = result.list
      trafficEventsTotal.value = result.total
      syncTrafficPagination(result.page, result.page_size)
    } finally {
      if (requestToken === trafficEventsRequestToken) {
        loadingTrafficEvents.value = false
      }
    }
  }

  return {
    services,
    attacks,
    summary,
    trafficSummary,
    trafficEvents,
    trafficEventsTotal,
    scoreboardRows,
    scoreboardFrozen,
    loadingRoundDetail,
    loadingTrafficSummary,
    loadingTrafficEvents,
    clearRoundDetail,
    refreshRoundDetail,
    refreshTrafficEvents,
  }
}
