import { type Ref } from 'vue'

import {
  getContestAWDReadiness,
  getContestAWDInstanceOrchestration,
  listContestTeams,
  listContestAWDRounds,
  listContestAWDServices,
} from '@/api/admin/contests'
import type {
  AWDRoundData,
  AdminContestChallengeData,
  AdminContestTeamData,
  ContestDetailData,
} from '@/api/contracts'
import { mapPlatformContestAwdServicesToChallengeLinks } from '@/utils/platformContestAwdChallengeLinks'

import {
  createEmptyInstanceOrchestration,
  loadStoredSelectedRoundId,
  pickRoundId,
} from './awdAdminSupport'

interface UseAwdContestSnapshotLoaderOptions {
  selectedContest: Readonly<Ref<ContestDetailData | null>>
  rounds: Ref<AWDRoundData[]>
  selectedRoundId: Ref<string | null>
  teams: Ref<AdminContestTeamData[]>
  challengeLinks: Ref<AdminContestChallengeData[]>
  instanceOrchestration: Ref<ReturnType<typeof createEmptyInstanceOrchestration>>
  readiness: Ref<Awaited<ReturnType<typeof getContestAWDReadiness>> | null>
  loadingRounds: Ref<boolean>
  loadingReadiness: Ref<boolean>
  loadingInstanceOrchestration: Ref<boolean>
  closeOverrideDialog: () => void
  clearRoundDetail: () => void
  refreshRoundDetail: (roundId?: string | null) => Promise<void>
  resetTrafficFiltersState: () => void
}

export function useAwdContestSnapshotLoader(options: UseAwdContestSnapshotLoaderOptions) {
  const {
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
  } = options

  let roundsRequestToken = 0
  let syncingSelectedRound = false

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
      const previousSelectedRound = rounds.value.find((item) => item.id === selectedRoundId.value) || null
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

  return {
    refresh,
    isSyncingSelectedRound: () => syncingSelectedRound,
  }
}
