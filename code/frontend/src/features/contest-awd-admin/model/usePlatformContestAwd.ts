import { computed, onBeforeUnmount, ref, watch, type Ref } from 'vue'

import {
  createContestAWDService,
  createContestAWDRound,
  createContestAWDAttackLog,
  createContestAWDServiceCheck,
  getContestAWDReadiness,
  getContestAWDInstanceOrchestration,
  getAdminContestLiveScoreboard,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  listContestAWDServices,
  listContestTeams,
  listContestAWDRoundAttacks,
  listContestAWDRoundTrafficEvents,
  listContestAWDRoundServices,
  listContestAWDRounds,
  runContestAWDRoundCheck,
  runContestAWDCurrentRoundCheck,
  updateContestAWDService,
} from '@/api/admin/contests'
import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficEventData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  AdminChallengeListItem,
  AdminContestChallengeData,
  AdminContestTeamData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import { mapPlatformContestAwdServicesToChallengeLinks } from '@/utils/platformContestAwdChallengeLinks'
import {
  createEmptyInstanceOrchestration,
  humanizeRequestError,
  isAWDReadinessBlockedError,
  loadStoredSelectedRoundId,
  persistSelectedRoundId,
  pickRoundId,
  type AWDTrafficFilterState,
} from './awdAdminSupport'
import { useAwdReadinessDecision } from './useAwdReadinessDecision'
import { useAwdServiceOperations } from './useAwdServiceOperations'
import { useAwdTrafficFilterState } from './useAwdTrafficFilterState'

const AWD_AUTO_REFRESH_INTERVAL_MS = 15_000

export function usePlatformContestAwd(selectedContest: Readonly<Ref<ContestDetailData | null>>) {
  const toast = useToast()
  const rounds = ref<AWDRoundData[]>([])
  const selectedRoundId = ref<string | null>(null)
  const services = ref<AWDTeamServiceData[]>([])
  const attacks = ref<AWDAttackLogData[]>([])
  const summary = ref<AWDRoundSummaryData | null>(null)
  const trafficSummary = ref<AWDTrafficSummaryData | null>(null)
  const trafficEvents = ref<AWDTrafficEventData[]>([])
  const trafficEventsTotal = ref(0)
  const {
    trafficFilters,
    buildTrafficEventsParams,
    applyTrafficFiltersPatch,
    setTrafficPageState,
    syncTrafficPagination,
    resetTrafficFiltersState,
  } = useAwdTrafficFilterState()
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

  const {
    instanceOrchestration,
    loadingInstanceOrchestration,
    startingInstanceKey,
    refreshInstanceOrchestration,
    startTeamServiceInstance,
    startTeamAllServices,
    startAllTeamServices,
  } = useAwdServiceOperations({
    selectedContest,
  })

  const {
    readiness,
    loadingReadiness,
    overrideDialogState,
    refreshReadiness,
    openOverrideDialog,
    closeOverrideDialog,
    confirmOverrideAction,
  } = useAwdReadinessDecision({
    selectedContest,
    onAfterOverride: async (preferredRoundId?: string) => {
      await refresh(preferredRoundId)
    },
  })

  const selectedRound = computed(
    () => rounds.value.find((item) => item.id === selectedRoundId.value) || null
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

  async function refreshChallengeLinks() {
    if (!selectedContest.value) {
      challengeLinks.value = []
      return
    }
    const nextServices = await listContestAWDServices(selectedContest.value.id)
    challengeLinks.value = mapPlatformContestAwdServicesToChallengeLinks(nextServices)
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

  async function applyTrafficFilters(
    patch: Partial<
      Pick<
        AWDTrafficFilterState,
        | 'attacker_team_id'
        | 'victim_team_id'
        | 'service_id'
        | 'awd_challenge_id'
        | 'status_group'
        | 'path_keyword'
      >
    >
  ) {
    applyTrafficFiltersPatch(patch)
    await refreshTrafficEvents(selectedRoundId.value)
  }

  async function setTrafficPage(page: number) {
    setTrafficPageState(page)
    await refreshTrafficEvents(selectedRoundId.value)
  }

  async function resetTrafficFilters() {
    resetTrafficFiltersState()
    await refreshTrafficEvents(selectedRoundId.value)
  }

  async function refresh(preferredRoundId?: string) {
    if (!selectedContest.value || selectedContest.value.mode !== 'awd') {
      rounds.value = []
      selectedRoundId.value = null
      resetTrafficFiltersState()
      teams.value = []
      challengeLinks.value = []
      instanceOrchestration.value = createEmptyInstanceOrchestration()
      readiness.value = null
      loadingReadiness.value = false
      loadingInstanceOrchestration.value = false
      closeOverrideDialog()
      clearRoundDetail()
      return
    }

    const requestToken = ++roundsRequestToken
    loadingRounds.value = true
    loadingReadiness.value = true
    loadingInstanceOrchestration.value = true
    try {
      const previousSelectedRound = selectedRound.value
      const wasFollowingRunningRound = previousSelectedRound?.status === 'running'
      const storedRoundId = loadStoredSelectedRoundId(selectedContest.value.id)
      const [
        nextRounds,
        nextTeams,
        nextContestAWDServices,
        nextInstanceOrchestration,
        nextReadiness,
      ] = await Promise.all([
        listContestAWDRounds(selectedContest.value.id),
        listContestTeams(selectedContest.value.id),
        listContestAWDServices(selectedContest.value.id),
        getContestAWDInstanceOrchestration(selectedContest.value.id),
        getContestAWDReadiness(selectedContest.value.id),
      ])
      if (requestToken !== roundsRequestToken) {
        return
      }

      rounds.value = nextRounds
      instanceOrchestration.value = nextInstanceOrchestration
      teams.value = nextTeams
      challengeLinks.value = mapPlatformContestAwdServicesToChallengeLinks(nextContestAWDServices)
      readiness.value = nextReadiness
      let nextPreferredRoundId = preferredRoundId || storedRoundId || undefined
      if (wasFollowingRunningRound) {
        const previousRoundStillRunning = nextRounds.some(
          (item) => item.id === previousSelectedRound?.id && item.status === 'running'
        )
        if (!previousRoundStillRunning) {
          nextPreferredRoundId =
            nextRounds.find((item) => item.status === 'running')?.id || nextPreferredRoundId
        }
      }
      syncingSelectedRound = true
      selectedRoundId.value = pickRoundId(nextRounds, selectedRoundId.value, nextPreferredRoundId)
      syncingSelectedRound = false
      await refreshRoundDetail(selectedRoundId.value)
    } finally {
      if (requestToken === roundsRequestToken) {
        loadingRounds.value = false
        loadingReadiness.value = false
        loadingInstanceOrchestration.value = false
      }
    }
  }

  async function loadChallengeCatalog() {
    challengeCatalog.value = []
  }

  async function runSelectedRoundCheck() {
    if (!selectedContest.value) {
      return
    }

    const activeRoundId = selectedRoundId.value
    const shouldRunCurrentRound = selectedRound.value?.status === 'running' || !activeRoundId
    checking.value = true
    try {
      let result
      if (shouldRunCurrentRound) {
        result = await runContestAWDCurrentRoundCheck(selectedContest.value.id)
      } else {
        result = await runContestAWDRoundCheck(selectedContest.value.id, activeRoundId)
      }
      toast.success(`第 ${result.round.round_number} 轮服务巡检已执行`)
      await refresh(result.round.id)
    } catch (error) {
      if (shouldRunCurrentRound && isAWDReadinessBlockedError(error)) {
        await openOverrideDialog('run_current_round_check', '立即巡检当前轮')
        return
      }
      toast.error(humanizeRequestError(error, '执行巡检失败'))
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
    } catch (error) {
      if (isAWDReadinessBlockedError(error)) {
        await openOverrideDialog('create_round', '创建轮次', payload)
        return
      }
      toast.error(humanizeRequestError(error, '创建轮次失败'))
    } finally {
      creatingRound.value = false
    }
  }

  async function createServiceCheck(payload: {
    team_id: number
    service_id: number
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
    service_id: number
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
    awd_challenge_id?: number
    points: number
    order?: number
    is_visible?: boolean
    awd_checker_type?: AdminContestChallengeData['awd_checker_type']
    awd_checker_config?: Record<string, unknown>
    awd_sla_score?: number
    awd_defense_score?: number
    awd_checker_preview_token?: string
  }) {
    if (!selectedContest.value) {
      return
    }
    if (!payload.awd_challenge_id) {
      toast.error('请选择 AWD 题目')
      return
    }

    savingChallengeConfig.value = true
    try {
      await createContestAWDService(selectedContest.value.id, {
        awd_challenge_id: payload.awd_challenge_id,
        points: payload.points,
        order: payload.order,
        is_visible: payload.is_visible,
        checker_type: payload.awd_checker_type,
        checker_config: payload.awd_checker_config,
        awd_sla_score: payload.awd_sla_score,
        awd_defense_score: payload.awd_defense_score,
        awd_checker_preview_token: payload.awd_checker_preview_token,
      })
      toast.success('赛事题目已关联')
      await refreshChallengeLinks()
      await refreshInstanceOrchestration()
      await refreshReadiness()
    } finally {
      savingChallengeConfig.value = false
    }
  }

  async function updateChallengeLink(
    challengeId: string,
    payload: {
      awd_challenge_id?: number
      points?: number
      order?: number
      is_visible?: boolean
      awd_checker_type?: AdminContestChallengeData['awd_checker_type']
      awd_checker_config?: Record<string, unknown>
      awd_sla_score?: number
      awd_defense_score?: number
      awd_checker_preview_token?: string
    }
  ) {
    if (!selectedContest.value) {
      return
    }

    savingChallengeConfig.value = true
    try {
      const currentChallenge = challengeLinks.value.find((item) => item.challenge_id === challengeId)
      const currentAWDChallengeID = Number(currentChallenge?.awd_challenge_id || 0) || undefined
      const awdChallengeID = payload.awd_challenge_id ?? currentAWDChallengeID
      const points = payload.points ?? currentChallenge?.points
      const order = payload.order ?? currentChallenge?.order
      const isVisible = payload.is_visible ?? currentChallenge?.is_visible

      if (awdChallengeID && points !== undefined) {
        const nextPayload = {
          awd_challenge_id: awdChallengeID,
          points,
          order,
          is_visible: isVisible,
          checker_type: payload.awd_checker_type,
          checker_config: payload.awd_checker_config,
          awd_sla_score: payload.awd_sla_score,
          awd_defense_score: payload.awd_defense_score,
          awd_checker_preview_token: payload.awd_checker_preview_token,
        }
        if (currentChallenge?.awd_service_id) {
          await updateContestAWDService(selectedContest.value.id, currentChallenge.awd_service_id, nextPayload)
        } else {
          await createContestAWDService(selectedContest.value.id, nextPayload)
        }
      }
      toast.success('题目配置已更新')
      await refreshChallengeLinks()
      await refreshInstanceOrchestration()
      await refreshReadiness()
    } finally {
      savingChallengeConfig.value = false
    }
  }

  watch(
    () => selectedContest.value?.id || null,
    async (nextContestId, previousContestId) => {
      if (nextContestId !== previousContestId) {
        resetTrafficFiltersState()
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
    instanceOrchestration,
    readiness,
    loadingRounds,
    loadingRoundDetail,
    loadingTrafficSummary,
    loadingTrafficEvents,
    loadingChallengeCatalog,
    loadingInstanceOrchestration,
    loadingReadiness,
    checking,
    creatingRound,
    savingServiceCheck,
    savingAttackLog,
    savingChallengeConfig,
    startingInstanceKey,
    overrideDialogState,
    hasSelectedContest,
    shouldAutoRefresh,
    refresh,
    refreshReadiness,
    refreshInstanceOrchestration,
    refreshRoundDetail,
    refreshTrafficEvents,
    applyTrafficFilters,
    setTrafficPage,
    resetTrafficFilters,
    runSelectedRoundCheck,
    startTeamServiceInstance,
    startTeamAllServices,
    startAllTeamServices,
    confirmOverrideAction,
    closeOverrideDialog,
    createRound,
    createServiceCheck,
    createAttackLog,
    loadChallengeCatalog,
    createChallengeLink,
    updateChallengeLink,
  }
}
