import { computed, onBeforeUnmount, ref, watch, type Ref } from 'vue'

import {
  createAdminContestChallenge,
  createContestAWDRound,
  createContestAWDAttackLog,
  createContestAWDServiceCheck,
  getChallenges,
  getAdminContestLiveScoreboard,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  listAdminContestChallenges,
  listContestTeams,
  listContestAWDRoundAttacks,
  listContestAWDRoundTrafficEvents,
  listContestAWDRoundServices,
  listContestAWDRounds,
  runContestAWDRoundCheck,
  runContestAWDCurrentRoundCheck,
  updateAdminContestChallenge,
} from '@/api/admin'
import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficEventData,
  AWDTrafficStatusGroup,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  AdminChallengeListItem,
  AdminContestChallengeData,
  AdminContestTeamData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'

const AWD_AUTO_REFRESH_INTERVAL_MS = 15_000
const AWD_TRAFFIC_DEFAULT_PAGE_SIZE = 20

interface AWDTrafficFilterState {
  attacker_team_id: string
  victim_team_id: string
  challenge_id: string
  status_group: 'all' | AWDTrafficStatusGroup
  path_keyword: string
  page: number
  page_size: number
}

function createDefaultTrafficFilters(): AWDTrafficFilterState {
  return {
    attacker_team_id: '',
    victim_team_id: '',
    challenge_id: '',
    status_group: 'all',
    path_keyword: '',
    page: 1,
    page_size: AWD_TRAFFIC_DEFAULT_PAGE_SIZE,
  }
}

function getSelectedRoundStorageKey(contestId: string): string {
  return `ctf_admin_awd_selected_round:${contestId}`
}

function loadStoredSelectedRoundId(contestId: string): string | null {
  if (typeof window === 'undefined') {
    return null
  }
  const value = window.sessionStorage.getItem(getSelectedRoundStorageKey(contestId))
  return value?.trim() || null
}

function persistSelectedRoundId(contestId: string, roundId: string | null): void {
  if (typeof window === 'undefined') {
    return
  }
  const storageKey = getSelectedRoundStorageKey(contestId)
  if (roundId) {
    window.sessionStorage.setItem(storageKey, roundId)
    return
  }
  window.sessionStorage.removeItem(storageKey)
}

function pickRoundId(rounds: AWDRoundData[], currentRoundId: string | null, preferredRoundId?: string): string | null {
  if (preferredRoundId && rounds.some((item) => item.id === preferredRoundId)) {
    return preferredRoundId
  }
  if (currentRoundId && rounds.some((item) => item.id === currentRoundId)) {
    return currentRoundId
  }
  const runningRound = rounds.find((item) => item.status === 'running')
  return runningRound?.id || rounds[rounds.length - 1]?.id || null
}

export function useAdminContestAWD(selectedContest: Readonly<Ref<ContestDetailData | null>>) {
  const toast = useToast()
  const rounds = ref<AWDRoundData[]>([])
  const selectedRoundId = ref<string | null>(null)
  const services = ref<AWDTeamServiceData[]>([])
  const attacks = ref<AWDAttackLogData[]>([])
  const summary = ref<AWDRoundSummaryData | null>(null)
  const trafficSummary = ref<AWDTrafficSummaryData | null>(null)
  const trafficEvents = ref<AWDTrafficEventData[]>([])
  const trafficEventsTotal = ref(0)
  const trafficFilters = ref<AWDTrafficFilterState>(createDefaultTrafficFilters())
  const scoreboardRows = ref<ScoreboardRow[]>([])
  const scoreboardFrozen = ref(false)
  const teams = ref<AdminContestTeamData[]>([])
  const challengeLinks = ref<AdminContestChallengeData[]>([])
  const challengeCatalog = ref<AdminChallengeListItem[]>([])
  const loadingRounds = ref(false)
  const loadingRoundDetail = ref(false)
  const loadingTrafficSummary = ref(false)
  const loadingTrafficEvents = ref(false)
  const loadingChallengeCatalog = ref(false)
  const checking = ref(false)
  const creatingRound = ref(false)
  const savingServiceCheck = ref(false)
  const savingAttackLog = ref(false)
  const savingChallengeConfig = ref(false)

  const selectedRound = computed(() =>
    rounds.value.find((item) => item.id === selectedRoundId.value) || null
  )

  const hasSelectedContest = computed(
    () => Boolean(selectedContest.value) && selectedContest.value?.mode === 'awd'
  )
  const shouldAutoRefresh = computed(() => {
    if (!selectedContest.value || selectedContest.value.mode !== 'awd') {
      return false
    }
    if (selectedContest.value.status !== 'running' && selectedContest.value.status !== 'frozen') {
      return false
    }
    return selectedRound.value?.status === 'running'
  })

  let roundsRequestToken = 0
  let roundDetailRequestToken = 0
  let trafficEventsRequestToken = 0
  let syncingSelectedRound = false
  let autoRefreshTimer: number | null = null

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

  function getTrafficEventsParams() {
    const filters = trafficFilters.value
    return {
      attacker_team_id: filters.attacker_team_id || undefined,
      victim_team_id: filters.victim_team_id || undefined,
      challenge_id: filters.challenge_id || undefined,
      status_group: filters.status_group === 'all' ? undefined : filters.status_group,
      path_keyword: filters.path_keyword.trim() || undefined,
      page: filters.page,
      page_size: filters.page_size,
    }
  }

  async function refreshChallengeLinks() {
    if (!selectedContest.value) {
      challengeLinks.value = []
      return
    }
    challengeLinks.value = await listAdminContestChallenges(selectedContest.value.id)
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
      const [nextServices, nextAttacks, nextSummary, nextTrafficSummary, nextTrafficEvents, nextScoreboard] =
        await Promise.all([
        listContestAWDRoundServices(selectedContest.value.id, roundId),
        listContestAWDRoundAttacks(selectedContest.value.id, roundId),
        getContestAWDRoundSummary(selectedContest.value.id, roundId),
        getContestAWDRoundTrafficSummary(selectedContest.value.id, roundId),
        listContestAWDRoundTrafficEvents(selectedContest.value.id, roundId, getTrafficEventsParams()),
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
      trafficFilters.value = {
        ...trafficFilters.value,
        page: nextTrafficEvents.page,
        page_size: nextTrafficEvents.page_size,
      }
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
        getTrafficEventsParams()
      )
      if (requestToken !== trafficEventsRequestToken) {
        return
      }
      trafficEvents.value = result.list
      trafficEventsTotal.value = result.total
      trafficFilters.value = {
        ...trafficFilters.value,
        page: result.page,
        page_size: result.page_size,
      }
    } finally {
      if (requestToken === trafficEventsRequestToken) {
        loadingTrafficEvents.value = false
      }
    }
  }

  async function applyTrafficFilters(
    patch: Partial<
      Pick<
        AWDTrafficFilterState,
        'attacker_team_id' | 'victim_team_id' | 'challenge_id' | 'status_group' | 'path_keyword'
      >
    >
  ) {
    trafficFilters.value = {
      ...trafficFilters.value,
      ...patch,
      page: 1,
    }
    await refreshTrafficEvents(selectedRoundId.value)
  }

  async function setTrafficPage(page: number) {
    const normalizedPage = Number.isFinite(page) && page > 0 ? Math.floor(page) : 1
    trafficFilters.value = {
      ...trafficFilters.value,
      page: normalizedPage,
    }
    await refreshTrafficEvents(selectedRoundId.value)
  }

  async function resetTrafficFilters() {
    trafficFilters.value = createDefaultTrafficFilters()
    await refreshTrafficEvents(selectedRoundId.value)
  }

  async function refresh(preferredRoundId?: string) {
    if (!selectedContest.value || selectedContest.value.mode !== 'awd') {
      rounds.value = []
      selectedRoundId.value = null
      trafficFilters.value = createDefaultTrafficFilters()
      teams.value = []
      challengeLinks.value = []
      clearRoundDetail()
      return
    }

    const requestToken = ++roundsRequestToken
    loadingRounds.value = true
    try {
      const previousSelectedRound = selectedRound.value
      const wasFollowingRunningRound = previousSelectedRound?.status === 'running'
      const storedRoundId = loadStoredSelectedRoundId(selectedContest.value.id)
      const [nextRounds, nextTeams, nextChallengeLinks] = await Promise.all([
        listContestAWDRounds(selectedContest.value.id),
        listContestTeams(selectedContest.value.id),
        listAdminContestChallenges(selectedContest.value.id),
      ])
      if (requestToken !== roundsRequestToken) {
        return
      }

      rounds.value = nextRounds
      teams.value = nextTeams
      challengeLinks.value = nextChallengeLinks
      let nextPreferredRoundId = preferredRoundId || storedRoundId || undefined
      if (wasFollowingRunningRound) {
        const previousRoundStillRunning = nextRounds.some(
          (item) => item.id === previousSelectedRound?.id && item.status === 'running'
        )
        if (!previousRoundStillRunning) {
          nextPreferredRoundId = nextRounds.find((item) => item.status === 'running')?.id || nextPreferredRoundId
        }
      }
      syncingSelectedRound = true
      selectedRoundId.value = pickRoundId(
        nextRounds,
        selectedRoundId.value,
        nextPreferredRoundId
      )
      syncingSelectedRound = false
      await refreshRoundDetail(selectedRoundId.value)
    } finally {
      if (requestToken === roundsRequestToken) {
        loadingRounds.value = false
      }
    }
  }

  async function loadChallengeCatalog() {
    if (loadingChallengeCatalog.value) {
      return
    }

    loadingChallengeCatalog.value = true
    try {
      const result = await getChallenges({ page: 1, page_size: 200 })
      challengeCatalog.value = result.list
    } finally {
      loadingChallengeCatalog.value = false
    }
  }

  async function runSelectedRoundCheck() {
    if (!selectedContest.value) {
      return
    }

    checking.value = true
    try {
      const result = selectedRoundId.value
        ? await runContestAWDRoundCheck(selectedContest.value.id, selectedRoundId.value)
        : await runContestAWDCurrentRoundCheck(selectedContest.value.id)
      toast.success(`第 ${result.round.round_number} 轮服务巡检已执行`)
      await refresh(result.round.id)
    } finally {
      checking.value = false
    }
  }

  async function createRound(payload: {
    round_number: number
    status?: AWDRoundData['status']
    attack_score?: number
    defense_score?: number
  }) {
    if (!selectedContest.value) {
      return
    }

    creatingRound.value = true
    try {
      const round = await createContestAWDRound(selectedContest.value.id, payload)
      toast.success(`第 ${round.round_number} 轮已创建`)
      await refresh(round.id)
      return round
    } finally {
      creatingRound.value = false
    }
  }

  async function createServiceCheck(payload: {
    team_id: number
    challenge_id: number
    service_status: AWDTeamServiceData['service_status']
    check_result?: Record<string, unknown>
  }) {
    if (!selectedContest.value || !selectedRoundId.value) {
      return
    }

    savingServiceCheck.value = true
    try {
      await createContestAWDServiceCheck(selectedContest.value.id, selectedRoundId.value, payload)
      toast.success('服务检查结果已记录')
      await refreshRoundDetail(selectedRoundId.value)
    } finally {
      savingServiceCheck.value = false
    }
  }

  async function createAttackLog(payload: {
    attacker_team_id: number
    victim_team_id: number
    challenge_id: number
    attack_type: AWDAttackLogData['attack_type']
    submitted_flag?: string
    is_success: boolean
  }) {
    if (!selectedContest.value || !selectedRoundId.value) {
      return
    }

    savingAttackLog.value = true
    try {
      await createContestAWDAttackLog(selectedContest.value.id, selectedRoundId.value, payload)
      toast.success('攻击日志已记录')
      await refreshRoundDetail(selectedRoundId.value)
    } finally {
      savingAttackLog.value = false
    }
  }

  async function createChallengeLink(payload: {
    challenge_id: number
    points: number
    order?: number
    is_visible?: boolean
    awd_checker_type?: AdminContestChallengeData['awd_checker_type']
    awd_checker_config?: Record<string, unknown>
    awd_sla_score?: number
    awd_defense_score?: number
  }) {
    if (!selectedContest.value) {
      return
    }

    savingChallengeConfig.value = true
    try {
      await createAdminContestChallenge(selectedContest.value.id, payload)
      toast.success('赛事题目已关联')
      await refreshChallengeLinks()
    } finally {
      savingChallengeConfig.value = false
    }
  }

  async function updateChallengeLink(
    challengeId: string,
    payload: {
      points?: number
      order?: number
      is_visible?: boolean
      awd_checker_type?: AdminContestChallengeData['awd_checker_type']
      awd_checker_config?: Record<string, unknown>
      awd_sla_score?: number
      awd_defense_score?: number
    }
  ) {
    if (!selectedContest.value) {
      return
    }

    savingChallengeConfig.value = true
    try {
      await updateAdminContestChallenge(selectedContest.value.id, challengeId, payload)
      toast.success('题目配置已更新')
      await refreshChallengeLinks()
    } finally {
      savingChallengeConfig.value = false
    }
  }

  watch(
    () => selectedContest.value?.id || null,
    async (nextContestId, previousContestId) => {
      if (nextContestId !== previousContestId) {
        trafficFilters.value = createDefaultTrafficFilters()
      }
      await refresh()
    },
    { immediate: true }
  )

  watch(
    () => selectedRoundId.value,
    async (nextRoundId, previousRoundId) => {
      if (!selectedContest.value || !nextRoundId || nextRoundId === previousRoundId) {
        if (!nextRoundId) {
          clearRoundDetail()
        }
        return
      }
      if (syncingSelectedRound) {
        return
      }
      await refreshRoundDetail(nextRoundId)
    }
  )

  watch(
    () => [selectedContest.value?.id || null, selectedRoundId.value] as const,
    ([contestId, roundId]) => {
      if (!contestId) {
        return
      }
      persistSelectedRoundId(contestId, roundId)
    },
    { immediate: true }
  )

  function stopAutoRefresh() {
    if (autoRefreshTimer !== null) {
      window.clearInterval(autoRefreshTimer)
      autoRefreshTimer = null
    }
  }

  watch(
    shouldAutoRefresh,
    (enabled) => {
      stopAutoRefresh()
      if (!enabled || typeof window === 'undefined') {
        return
      }
      autoRefreshTimer = window.setInterval(() => {
        void refresh()
      }, AWD_AUTO_REFRESH_INTERVAL_MS)
    },
    { immediate: true }
  )

  onBeforeUnmount(() => {
    stopAutoRefresh()
  })

  return {
    rounds,
    selectedRoundId,
    selectedRound,
    services,
    attacks,
    summary,
    trafficSummary,
    trafficEvents,
    trafficEventsTotal,
    trafficFilters,
    scoreboardRows,
    scoreboardFrozen,
    teams,
    challengeLinks,
    challengeCatalog,
    loadingRounds,
    loadingRoundDetail,
    loadingTrafficSummary,
    loadingTrafficEvents,
    loadingChallengeCatalog,
    checking,
    creatingRound,
    savingServiceCheck,
    savingAttackLog,
    savingChallengeConfig,
    hasSelectedContest,
    shouldAutoRefresh,
    refresh,
    refreshRoundDetail,
    refreshTrafficEvents,
    applyTrafficFilters,
    setTrafficPage,
    resetTrafficFilters,
    runSelectedRoundCheck,
    createRound,
    createServiceCheck,
    createAttackLog,
    loadChallengeCatalog,
    createChallengeLink,
    updateChallengeLink,
  }
}
