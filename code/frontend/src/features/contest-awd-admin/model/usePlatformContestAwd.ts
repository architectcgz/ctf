import { computed, ref, type Ref } from 'vue'

import type {
  AWDRoundData,
  AdminChallengeListItem,
  AdminContestChallengeData,
  AdminContestTeamData,
  ContestDetailData,
} from '@/api/contracts'
import {
  createEmptyInstanceOrchestration,
} from './awdAdminSupport'
import { useAwdContestStateFlags } from './useAwdContestStateFlags'
import { useAwdReadinessDecision } from './useAwdReadinessDecision'
import { useAwdChallengeLinkOperations } from './useAwdChallengeLinkOperations'
import { useAwdContestSnapshotLoader } from './useAwdContestSnapshotLoader'
import { useAwdLifecycleBindings } from './useAwdLifecycleBindings'
import { useAwdRoundDetailState } from './useAwdRoundDetailState'
import { useAwdRoundOperations } from './useAwdRoundOperations'
import { useAwdServiceOperations } from './useAwdServiceOperations'
import { useAwdTrafficActions } from './useAwdTrafficActions'
import { useAwdTrafficFilterState } from './useAwdTrafficFilterState'

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
  const { hasSelectedContest, shouldAutoRefresh } = useAwdContestStateFlags({
    selectedContest,
    selectedRound,
  })
  const { applyTrafficFilters, setTrafficPage, resetTrafficFilters } = useAwdTrafficActions({
    selectedRoundId,
    applyTrafficFiltersPatch,
    setTrafficPageState,
    resetTrafficFiltersState,
    refreshTrafficEvents,
  })
  const { refresh, isSyncingSelectedRound } = useAwdContestSnapshotLoader({
    selectedContest,
    rounds,
    selectedRoundId,
    teams,
    challengeLinks,
    instanceOrchestration,
    readiness,
    loadingRounds,
    loadingReadiness,
    loadingInstanceOrchestration,
    closeOverrideDialog,
    clearRoundDetail,
    refreshRoundDetail,
    resetTrafficFiltersState,
  })

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

  useAwdLifecycleBindings({
    selectedContest,
    selectedRoundId,
    shouldAutoRefresh,
    refresh: async () => {
      await refresh()
    },
    refreshRoundDetail: async (roundId) => {
      await refreshRoundDetail(roundId)
    },
    clearRoundDetail,
    resetTrafficFiltersState,
    isSyncingSelectedRound,
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
