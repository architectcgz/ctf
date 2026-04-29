import type {
  AdminContestAWDServiceData,
  AdminContestChallengeData,
} from '@/api/contracts'

function normalizeCheckerConfig(
  service?: AdminContestAWDServiceData
): NonNullable<AdminContestChallengeData['awd_checker_config']> {
  if (service?.checker_config && typeof service.checker_config === 'object') {
    return service.checker_config
  }
  return {}
}

export function mapPlatformContestAwdServicesToChallengeLinks(
  services: AdminContestAWDServiceData[]
): AdminContestChallengeData[] {
  return services.map((service) => ({
    id: service.id,
    contest_id: service.contest_id,
    challenge_id: service.awd_challenge_id,
    awd_challenge_id: service.awd_challenge_id,
    title: service.title || service.display_name,
    category: service.category,
    difficulty: service.difficulty,
    points: normalizeAwdServicePoints(service),
    order: service.order,
    is_visible: service.is_visible,
    created_at: service.created_at,
    awd_service_id: service.id,
    awd_service_display_name: service.display_name || undefined,
    awd_checker_type: service.checker_type,
    awd_checker_config: normalizeCheckerConfig(service),
    awd_sla_score: service.sla_score ?? 0,
    awd_defense_score: service.defense_score ?? 0,
    awd_checker_validation_state: service.validation_state || 'pending',
    awd_checker_last_preview_at: service.last_preview_at,
    awd_checker_last_preview_result: service.last_preview_result,
  }))
}

function normalizeAwdServicePoints(service: AdminContestAWDServiceData): number {
  const value = service.score_config?.points
  return typeof value === 'number' && Number.isFinite(value) ? value : 0
}
