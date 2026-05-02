import { computed, onBeforeUnmount, ref, watch, type Ref } from 'vue'

import {
  getContestAWDReadiness,
  getContestAWDInstanceOrchestration,
  listContestTeams,
  listContestAWDRounds,
  listContestAWDServices,
} from '@/api/admin/contests'
import type {
  AWDRoundData,
  AdminChallengeListItem,
  AdminContestChallengeData,
  AdminContestTeamData,
  ContestDetailData,
} from '@/api/contracts'
import { mapPlatformContestAwdServicesToChallengeLinks } from '@/utils/platformContestAwdChallengeLinks'
import {
  createEmptyInstanceOrchestration,
  loadStoredSelectedRoundId,
  persistSelectedRoundId,
  pickRoundId,
  type AWDTrafficFilterState,
} from './awdAdminSupport'
import { useAwdReadinessDecision } from './useAwdReadinessDecision'
import { useAwdChallengeLinkOperations } from './useAwdChallengeLinkOperations'
import { useAwdRoundDetailState } from './useAwdRoundDetailState'
import { useAwdRoundOperations } from './useAwdRoundOperations'
import { useAwdServiceOperations } from './useAwdServiceOperations'
import { useAwdTrafficFilterState } from './useAwdTrafficFilterState'

const AWD_AUTO_REFRESH_INTERVAL_MS = 15_000

export function usePlatformContestAwd(selectedContest: Readonly<Ref<ContestDetailData | null>>) {
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

  const {
    checking,
    creatingRound,
    savingServiceCheck,
    savingAttackLog,
    runSelectedRoundCheck,
    createRound,
    createServiceCheck,
    createAttackLog,
  } = useAwdRoundOperations({
    selectedContest,
    selectedRoundId,
    selectedRound,
    refresh,
    refreshRoundDetail,
    openOverrideDialog,
  })

  async function loadChallengeCatalog() {
    challengeCatalog.value = []
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
