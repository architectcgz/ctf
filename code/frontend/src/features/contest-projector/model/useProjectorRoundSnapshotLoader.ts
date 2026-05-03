import type { Ref } from 'vue'

import {
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  listContestAWDRoundAttacks,
  listContestAWDRoundServices,
} from '@/api/admin/contests'
import type {
  AWDAttackLogData,
  AWDRoundSummaryData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
} from '@/api/contracts'

interface UseProjectorRoundSnapshotLoaderOptions {
  services: Ref<AWDTeamServiceData[]>
  attacks: Ref<AWDAttackLogData[]>
  roundSummary: Ref<AWDRoundSummaryData | null>
  trafficSummary: Ref<AWDTrafficSummaryData | null>
}

export function useProjectorRoundSnapshotLoader({
  services,
  attacks,
  roundSummary,
  trafficSummary,
}: UseProjectorRoundSnapshotLoaderOptions) {
  async function loadRoundSnapshot(
    contestId: string,
    roundId: string,
    requestToken: number,
    isStaleRequest: (token: number) => boolean
  ): Promise<void> {
    const [nextServices, nextAttacks, nextRoundSummary, nextTrafficSummary] = await Promise.all([
      listContestAWDRoundServices(contestId, roundId),
      listContestAWDRoundAttacks(contestId, roundId),
      getContestAWDRoundSummary(contestId, roundId),
      getContestAWDRoundTrafficSummary(contestId, roundId),
    ])
    if (isStaleRequest(requestToken)) {
      return
    }
    services.value = nextServices
    attacks.value = nextAttacks
    roundSummary.value = nextRoundSummary
    trafficSummary.value = nextTrafficSummary
  }

  function clearRoundSnapshot(): void {
    services.value = []
    attacks.value = []
    roundSummary.value = null
    trafficSummary.value = null
  }

  return {
    loadRoundSnapshot,
    clearRoundSnapshot,
  }
}
