import type { AWDAttackLogData, AWDTrafficEventData, AWDTrafficSummaryData, AWDTeamServiceData } from '@/api/contracts'

interface TeamOption {
  id: string
  name: string
}

export function buildAttackTeamOptions(attacks: AWDAttackLogData[]): TeamOption[] {
  const seen = new Set<string>()
  return attacks.flatMap((item) => {
    const entries = [
      { id: item.attacker_team_id, name: item.attacker_team },
      { id: item.victim_team_id, name: item.victim_team },
    ]
    return entries.filter((entry) => {
      if (seen.has(entry.id)) {
        return false
      }
      seen.add(entry.id)
      return true
    })
  })
}

export function buildAttackSourceOptions(attacks: AWDAttackLogData[]): AWDAttackLogData['source'][] {
  const seen = new Set<AWDAttackLogData['source']>()
  return attacks
    .map((item) => item.source)
    .filter((item) => {
      if (seen.has(item)) {
        return false
      }
      seen.add(item)
      return true
    })
}

interface BuildTrafficTeamOptionsOptions {
  services: AWDTeamServiceData[]
  attackTeamOptions: TeamOption[]
  trafficSummary: AWDTrafficSummaryData | null
  trafficEvents: AWDTrafficEventData[]
}

export function buildTrafficTeamOptions({
  services,
  attackTeamOptions,
  trafficSummary,
  trafficEvents,
}: BuildTrafficTeamOptionsOptions): TeamOption[] {
  const entries = new Map<string, string>()

  for (const service of services) {
    entries.set(service.team_id, service.team_name)
  }
  for (const attackTeam of attackTeamOptions) {
    entries.set(attackTeam.id, attackTeam.name)
  }
  for (const item of trafficSummary?.top_attackers || []) {
    entries.set(item.team_id, item.team_name)
  }
  for (const item of trafficSummary?.top_victims || []) {
    entries.set(item.team_id, item.team_name)
  }
  for (const item of trafficEvents) {
    if (item.attacker_team_name?.trim()) {
      entries.set(item.attacker_team_id, item.attacker_team_name)
    }
    if (item.victim_team_name?.trim()) {
      entries.set(item.victim_team_id, item.victim_team_name)
    }
  }

  return [...entries.entries()]
    .map(([id, name]) => ({ id, name }))
    .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
}

interface AttackFilterOptions {
  attackTeamFilter: string
  attackResultFilter: 'all' | 'success' | 'failed'
  attackSourceFilter: 'all' | AWDAttackLogData['source']
}

export function filterAttacks(
  attacks: AWDAttackLogData[],
  { attackTeamFilter, attackResultFilter, attackSourceFilter }: AttackFilterOptions
): AWDAttackLogData[] {
  return attacks.filter((item) => {
    if (
      attackTeamFilter &&
      item.attacker_team_id !== attackTeamFilter &&
      item.victim_team_id !== attackTeamFilter
    ) {
      return false
    }
    if (attackResultFilter === 'success' && !item.is_success) {
      return false
    }
    if (attackResultFilter === 'failed' && item.is_success) {
      return false
    }
    if (attackSourceFilter !== 'all' && item.source !== attackSourceFilter) {
      return false
    }
    return true
  })
}
