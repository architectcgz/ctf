import { computed, onBeforeUnmount, ref, watch, type Ref } from 'vue'

import {
  createContestAWDRound,
  createContestAWDAttackLog,
  createContestAWDServiceCheck,
  getContestAWDReadiness,
  getContestAWDInstanceOrchestration,
  listContestTeams,
  listContestAWDRounds,
  runContestAWDRoundCheck,
  runContestAWDCurrentRoundCheck,
  listContestAWDServices,
} from '@/api/admin/contests'
import type {
  AWDAttackLogData,
  AWDRoundData,
  AWDTeamServiceData,
  AdminChallengeListItem,
  AdminContestChallengeData,
  AdminContestTeamData,
  ContestDetailData,
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
import { useAwdChallengeLinkOperations } from './useAwdChallengeLinkOperations'
import { useAwdRoundDetailState } from './useAwdRoundDetailState'
import { useAwdServiceOperations } from './useAwdServiceOperations'
import { useAwdTrafficFilterState } from './useAwdTrafficFilterState'

const AWD_AUTO_REFRESH_INTERVAL_MS = 15_000

export function usePlatformContestAwd(selectedContest: Readonly<Ref<ContestDetailData | null>>) {
  const toast = useToast()
  const rounds = ref<AWDRoundData[]>([])
  const selectedRoundId = ref<string | null>(null)
  const {
    trafficFilters,
    buildTrafficEventsParams,
    applyTrafficFiltersPatch,
    setTrafficPageState,
    syncTrafficPagination,
    resetTrafficFiltersState,
  } = useAwdTrafficFilterState()
  const teams = ref<AdminContestTeamData[]>([])
  const challengeCatalog = ref<AdminChallengeListItem[]>([])
  const loadingRounds = ref(false)
  const loadingChallengeCatalog = ref(false)
  const checking = ref(false)
  const creatingRound = ref(false)
  const savingServiceCheck = ref(false)
  const savingAttackLog = ref(false)

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
    challengeLinks,
    savingChallengeConfig,
    refreshChallengeLinks,
    createChallengeLink,
    updateChallengeLink,
  } = useAwdChallengeLinkOperations({
    selectedContest,
    onAfterMutate: async () => {
      await refreshInstanceOrchestration()
      await refreshReadiness()
    },
  })
  const {
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
  } = useAwdRoundDetailState({
    selectedContest,
    selectedRoundId,
    buildTrafficEventsParams,
    syncTrafficPagination,
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
  let syncingSelectedRound = false
  let autoRefreshTimer: number | null = null

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
