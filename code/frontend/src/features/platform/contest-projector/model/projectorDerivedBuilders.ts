import type { AWDAttackLogData, AWDTeamServiceData, AWDTrafficSummaryData } from '@/api/contracts'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorAttackLeader,
  ContestProjectorServiceMatrixRow,
  ContestProjectorTrafficTrendBar,
} from './projectorTypes'

export function buildServiceMatrixRows(services: AWDTeamServiceData[]): ContestProjectorServiceMatrixRow[] {
  const teamMap = new Map<string, ContestProjectorServiceMatrixRow>()
  for (const service of services) {
    const row = teamMap.get(service.team_id) ?? {
      team_id: service.team_id,
      team_name: service.team_name,
      services: [],
    }
    row.services.push(service)
    teamMap.set(service.team_id, row)
  }
  return Array.from(teamMap.values()).slice(0, 10)
}

export function buildAttackLeaders(attacks: AWDAttackLogData[]): ContestProjectorAttackLeader[] {
  const teamMap = new Map<string, ContestProjectorAttackLeader>()
  for (const attack of attacks) {
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
}

export function buildAttackEdges(
  attacks: AWDAttackLogData[],
  services: AWDTeamServiceData[]
): ContestProjectorAttackEdge[] {
  const serviceLabelMap = new Map<string, string>()
  for (const service of services) {
    const label =
      service.service_name?.trim() ||
      service.awd_challenge_title?.trim() ||
      (service.service_id ? `服务 ${service.service_id}` : `题目 ${service.awd_challenge_id}`)
    if (service.service_id) {
      serviceLabelMap.set(`${service.team_id}:service:${service.service_id}`, label)
    }
    serviceLabelMap.set(`${service.team_id}:challenge:${service.awd_challenge_id}`, label)
  }

  const edgeMap = new Map<string, ContestProjectorAttackEdge>()
  for (const attack of attacks) {
    const edgeId = `${attack.attacker_team_id}->${attack.victim_team_id}`
    const current = edgeMap.get(edgeId)
    const attackTime = new Date(attack.created_at).getTime()
    const currentTime = current ? new Date(current.latest_at).getTime() : 0
    const latestServiceLabel =
      (attack.service_id
        ? serviceLabelMap.get(`${attack.victim_team_id}:service:${attack.service_id}`)
        : undefined) ??
      serviceLabelMap.get(`${attack.victim_team_id}:challenge:${attack.awd_challenge_id}`) ??
      (attack.service_id ? `服务 ${attack.service_id}` : `题目 ${attack.awd_challenge_id}`)
    const next: ContestProjectorAttackEdge = current ?? {
      id: edgeId,
      attacker_team_id: attack.attacker_team_id,
      attacker_team: attack.attacker_team,
      victim_team_id: attack.victim_team_id,
      victim_team: attack.victim_team,
      latest_service_id: attack.service_id,
      latest_challenge_id: attack.awd_challenge_id,
      latest_target_key: attack.service_id
        ? `${attack.victim_team_id}:service:${attack.service_id}`
        : `${attack.victim_team_id}:challenge:${attack.awd_challenge_id}`,
      success: 0,
      failed: 0,
      total: 0,
      score: 0,
      latest_at: attack.created_at,
      latest_service_label: latestServiceLabel,
      successRate: 0,
      reciprocalSuccess: 0,
    }
    next.total += 1
    if (attack.is_success) {
      next.success += 1
      next.score += attack.score_gained
    } else {
      next.failed += 1
    }
    if (attackTime >= currentTime) {
      next.latest_at = attack.created_at
      next.latest_service_label = latestServiceLabel
      next.latest_service_id = attack.service_id
      next.latest_challenge_id = attack.awd_challenge_id
      next.latest_target_key = attack.service_id
        ? `${attack.victim_team_id}:service:${attack.service_id}`
        : `${attack.victim_team_id}:challenge:${attack.awd_challenge_id}`
    }
    next.successRate = Math.round((next.success / Math.max(next.total, 1)) * 100)
    edgeMap.set(edgeId, next)
  }

  const edges = Array.from(edgeMap.values())
  for (const edge of edges) {
    edge.reciprocalSuccess = edgeMap.get(`${edge.victim_team_id}->${edge.attacker_team_id}`)?.success ?? 0
  }

  return edges
    .sort(
      (a, b) =>
        b.success - a.success ||
        b.score - a.score ||
        new Date(b.latest_at).getTime() - new Date(a.latest_at).getTime()
    )
    .slice(0, 8)
}

export function buildTrafficTrendBars(
  trafficSummary: AWDTrafficSummaryData | null
): ContestProjectorTrafficTrendBar[] {
  const buckets = (trafficSummary?.trend_buckets ?? []).slice(-12)
  const maxRequests = Math.max(...buckets.map((item) => item.request_count), 1)
  return buckets.map((item) => ({
    bucket_start_at: item.bucket_start_at,
    request_count: item.request_count,
    error_count: item.error_count,
    height: `${Math.max(10, Math.round((item.request_count / maxRequests) * 100))}%`,
    errorHeight: `${Math.min(100, Math.round((item.error_count / Math.max(item.request_count, 1)) * 100))}%`,
  }))
}
