import type { ChallengeCategory, ChallengeDifficulty } from '@/entities/challenge/model'

type IDLike = string
type AWDCheckerType = 'legacy_probe' | 'http_standard' | 'tcp_standard' | 'script_checker'
type AWDCheckerValidationState = 'pending' | 'passed' | 'failed' | 'stale'
type AWDServiceStatus = 'up' | 'down' | 'compromised'

interface AwdCheckerPreviewContext {
  access_url: string
  preview_flag: string
  round_number: number
  team_id: IDLike
  awd_challenge_id: IDLike
}

interface AwdCheckerPreviewData {
  checker_type?: AWDCheckerType
  service_status: AWDServiceStatus
  check_result: Record<string, unknown>
  preview_context: AwdCheckerPreviewContext
  preview_token?: string
}

interface ContestAwdServiceLinkSource {
  id: IDLike
  contest_id: IDLike
  awd_challenge_id: IDLike
  title?: string
  category?: ChallengeCategory
  difficulty?: ChallengeDifficulty
  display_name: string
  order: number
  is_visible: boolean
  score_config?: Record<string, unknown>
  checker_type?: AWDCheckerType
  checker_config?: Record<string, unknown>
  sla_score?: number
  defense_score?: number
  validation_state?: AWDCheckerValidationState
  last_preview_at?: string
  last_preview_result?: AwdCheckerPreviewData
  created_at: string
}

export interface ContestAwdChallengeLink {
  id: IDLike
  contest_id: IDLike
  challenge_id: IDLike
  awd_challenge_id?: IDLike
  title?: string
  category?: ChallengeCategory
  difficulty?: ChallengeDifficulty
  points: number
  order: number
  is_visible: boolean
  created_at: string
  awd_service_id?: IDLike
  awd_service_display_name?: string
  awd_checker_type?: AWDCheckerType
  awd_checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_validation_state?: AWDCheckerValidationState
  awd_checker_last_preview_at?: string
  awd_checker_last_preview_result?: AwdCheckerPreviewData
}

function normalizeCheckerConfig(
  service?: ContestAwdServiceLinkSource
): NonNullable<ContestAwdChallengeLink['awd_checker_config']> {
  if (service?.checker_config && typeof service.checker_config === 'object') {
    return service.checker_config
  }
  return {}
}

export function mapPlatformContestAwdServicesToChallengeLinks(
  services: ContestAwdServiceLinkSource[]
): ContestAwdChallengeLink[] {
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

function normalizeAwdServicePoints(service: ContestAwdServiceLinkSource): number {
  const value = service.score_config?.points
  return typeof value === 'number' && Number.isFinite(value) ? value : 0
}
