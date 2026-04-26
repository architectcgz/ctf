import { computed, type Ref } from 'vue'

import type {
  AWDAttackLogData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  ScoreboardRow,
} from '@/api/contracts'
import type {
  ContestProjectorAttackLeader,
  ContestProjectorServiceMatrixRow,
  ContestProjectorTrafficTrendBar,
} from '@/components/platform/contest/projector/contestProjectorTypes'

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
  const serviceMatrixRows = computed<ContestProjectorServiceMatrixRow[]>(() => {
    const teamMap = new Map<string, ContestProjectorServiceMatrixRow>()
    for (const service of services.value) {
      const row = teamMap.get(service.team_id) ?? {
        team_id: service.team_id,
        team_name: service.team_name,
        services: [],
      }
      row.services.push(service)
      teamMap.set(service.team_id, row)
    }
    return Array.from(teamMap.values()).slice(0, 10)
  })
  const attackLeaders = computed<ContestProjectorAttackLeader[]>(() => {
    const teamMap = new Map<string, ContestProjectorAttackLeader>()
    for (const attack of attacks.value) {
      const row = teamMap.get(attack.attacker_team_id) ?? {
        team_id: attack.attacker_team_id,
        team_name: attack.attacker_team,
        success: 0,
        score: 0,
      }
      if (attack.is_success) {
        row.success += 1
        row.score += attack.score_gained
      }
      teamMap.set(attack.attacker_team_id, row)
    }
    return Array.from(teamMap.values())
      .sort((a, b) => b.success - a.success || b.score - a.score)
      .slice(0, 5)
  })
  const trafficTrendBars = computed<ContestProjectorTrafficTrendBar[]>(() => {
    const buckets = (trafficSummary.value?.trend_buckets ?? []).slice(-12)
    const maxRequests = Math.max(...buckets.map((item) => item.request_count), 1)
    return buckets.map((item) => ({
      bucket_start_at: item.bucket_start_at,
      request_count: item.request_count,
      error_count: item.error_count,
      height: `${Math.max(10, Math.round((item.request_count / maxRequests) * 100))}%`,
      errorHeight: `${Math.min(100, Math.round((item.error_count / Math.max(item.request_count, 1)) * 100))}%`,
    }))
  })
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
    trafficTrendBars,
    hotVictims,
  }
}
