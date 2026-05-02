import { computed, type Ref } from 'vue'

import type {
  AWDAttackLogData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  ScoreboardRow,
} from '@/api/contracts'
import {
  buildAttackEdges,
  buildAttackLeaders,
  buildServiceMatrixRows,
  buildTrafficTrendBars,
} from './projectorDerivedBuilders'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorAttackLeader,
  ContestProjectorServiceMatrixRow,
} from './projectorTypes'

interface UseContestProjectorDerivedOptions {
  scoreboardRows: Readonly<Ref<ScoreboardRow[]>>
  services: Readonly<Ref<AWDTeamServiceData[]>>
  attacks: Readonly<Ref<AWDAttackLogData[]>>
  trafficSummary: Readonly<Ref<AWDTrafficSummaryData | null>>
}

export function useContestProjectorDerived({
  scoreboardRows,
  services,
  attacks,
  trafficSummary,
}: UseContestProjectorDerivedOptions) {
  const topThreeRows = computed(() => scoreboardRows.value.slice(0, 3))
  const leaderboardRows = computed(() => scoreboardRows.value.slice(0, 10))
  const firstBlood = computed(() =>
    attacks.value
      .filter((item) => item.is_success)
      .slice()
      .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())[0] ?? null
  )
  const latestAttackEvents = computed(() =>
    attacks.value
      .slice()
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 8)
  )
  const serviceStatusCounts = computed(() => ({
    up: services.value.filter((item) => item.service_status === 'up').length,
    down: services.value.filter((item) => item.service_status === 'down').length,
    compromised: services.value.filter((item) => item.service_status === 'compromised').length,
  }))
  const serviceHealthRate = computed(() => {
    const total = services.value.length
    if (total === 0) return 0
    return Math.round((serviceStatusCounts.value.up / total) * 100)
  })
  const serviceMatrixRows = computed<ContestProjectorServiceMatrixRow[]>(() =>
    buildServiceMatrixRows(services.value)
  )
  const attackLeaders = computed<ContestProjectorAttackLeader[]>(() => buildAttackLeaders(attacks.value))
  const attackEdges = computed<ContestProjectorAttackEdge[]>(() =>
    buildAttackEdges(attacks.value, services.value)
  )
  const trafficTrendBars = computed(() => buildTrafficTrendBars(trafficSummary.value))
  const hotVictims = computed(() => (trafficSummary.value?.top_victims ?? []).slice(0, 4))

  return {
    topThreeRows,
    leaderboardRows,
    firstBlood,
    latestAttackEvents,
    serviceStatusCounts,
    serviceHealthRate,
    serviceMatrixRows,
    attackLeaders,
    attackEdges,
    trafficTrendBars,
    hotVictims,
  }
}
