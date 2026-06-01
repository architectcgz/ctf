import { normalizeInstanceData, type RawInstanceData } from '../instance'

import type {
  AWDAttackLogData,
  AWDCheckerPreviewData,
  AWDCheckerType,
  AWDReadinessAction,
  AWDReadinessBlockingReason,
  AWDReadinessData,
  AWDReadinessGlobalReason,
  AWDReadinessItemData,
  AWDRoundData,
  AWDRoundMetricsData,
  AWDRoundSummaryData,
  AWDRoundSummaryItemData,
  AWDTeamServiceData,
  AWDTrafficEventData,
  AWDTrafficEventPageData,
  AWDTrafficStatusGroup,
  AWDTrafficSummaryData,
  AWDTrafficTopChallengeData,
  AWDTrafficTopPathData,
  AWDTrafficTopTeamData,
  AWDTrafficTrendBucketData,
  AdminContestAWDInstanceItemData,
  AdminContestAWDInstanceOrchestrationData,
  AdminContestAWDInstancePrewarmData,
  AdminContestAWDInstancePrewarmItemData,
  AdminContestAWDInstanceServiceData,
  AdminContestAWDInstanceTeamData,
  AdminContestAWDServiceData,
  AdminContestChallengeRelationData,
  AdminContestTeamData,
  AWDCheckerRunData,
  ContestAnnouncement,
  ContestDetailData,
  ContestMode,
  ContestPageData,
  ContestScoreboardData,
  ContestStatus,
  PageResult,
} from '../contracts'

// Re-export contract types needed by consuming modules
export type {
  AWDCheckerRunData,
  AWDCheckerPreviewData,
  AWDAttackLogData,
  AWDRoundData,
  AWDRoundMetricsData,
  AWDRoundSummaryData,
  AWDRoundSummaryItemData,
  AWDReadinessData,
  AWDReadinessItemData,
  AWDTeamServiceData,
  AWDTrafficEventData,
  AWDTrafficEventPageData,
  AWDTrafficSummaryData,
  AWDTrafficTopChallengeData,
  AWDTrafficTopPathData,
  AWDTrafficTopTeamData,
  AWDTrafficTrendBucketData,
  AdminContestAWDInstanceItemData,
  AdminContestAWDInstanceOrchestrationData,
  AdminContestAWDInstancePrewarmData,
  AdminContestAWDInstancePrewarmItemData,
  AdminContestAWDInstanceServiceData,
  AdminContestAWDInstanceTeamData,
  AdminContestAWDServiceData,
  AdminContestChallengeRelationData,
  AdminContestTeamData,
  ContestAnnouncement,
  ContestDetailData,
  ContestPageData,
  ContestScoreboardData,
  PageResult,
} from '../contracts'

/* ── 类型别名 ── */

export type AdminContestStatus = Extract<
  ContestStatus,
  'draft' | 'registering' | 'running' | 'frozen' | 'ended'
>

export type AdminContestMode = Extract<ContestMode, 'jeopardy' | 'awd'>

/* ── Payload 接口 ── */

export interface AdminContestChallengeCreatePayload {
  challenge_id: number
  points: number
  order?: number
  is_visible?: boolean
}

export interface AdminContestChallengeUpdatePayload {
  points?: number
  order?: number
  is_visible?: boolean
}

export interface AdminContestAWDServiceCreatePayload {
  awd_challenge_id: number
  points: number
  display_name?: string
  order?: number
  is_visible?: boolean
  checker_type?: AWDCheckerType
  checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}

export interface AdminContestAWDServiceUpdatePayload {
  awd_challenge_id?: number
  points?: number
  display_name?: string
  order?: number
  is_visible?: boolean
  checker_type?: AWDCheckerType
  checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}

export interface AdminAWDCheckerPreviewPayload {
  service_id?: number
  awd_challenge_id: number
  checker_type: AWDCheckerType
  checker_config: Record<string, unknown>
  access_url?: string
  preview_flag?: string
  preview_request_id?: string
}

export interface AdminContestCreatePayload {
  title: string
  description?: string
  mode: AdminContestMode
  starts_at: string
  ends_at: string
}

export interface AdminContestUpdatePayload {
  title?: string
  description?: string
  mode?: AdminContestMode
  starts_at?: string
  ends_at?: string
  status?: AdminContestStatus
  force_override?: boolean
  override_reason?: string
}

export interface AdminContestAnnouncementCreatePayload {
  title: string
  content: string
}

export interface AdminAWDRoundCreatePayload {
  round_number: number
  status?: AWDRoundData['status']
  attack_score?: number
  defense_score?: number
  force_override?: boolean
  override_reason?: string
}

export interface AdminAWDCurrentRoundCheckPayload {
  force_override?: boolean
  override_reason?: string
}

export interface AdminAWDServiceCheckPayload {
  team_id: number
  service_id: number
  service_status: AWDTeamServiceData['service_status']
  check_result?: Record<string, unknown>
}

export interface AdminAWDAttackLogPayload {
  attacker_team_id: number
  victim_team_id: number
  service_id: number
  attack_type: AWDAttackLogData['attack_type']
  submitted_flag?: string
  is_success: boolean
}

export interface AdminAWDTrafficEventsParams {
  attacker_team_id?: string
  victim_team_id?: string
  service_id?: string
  awd_challenge_id?: string
  status_group?: AWDTrafficStatusGroup
  path_keyword?: string
  page?: number
  page_size?: number
}

/* ── Raw 接口 ── */

export interface RawContestItem {
  id: number
  title: string
  description: string
  mode: AdminContestMode
  start_time: string
  end_time: string
  freeze_time?: string | null
  status: 'draft' | 'registration' | 'running' | 'frozen' | 'ended'
  created_at: string
  updated_at: string
}

export interface RawContestListSummary {
  draft_count?: number
  registering_count?: number
  running_count?: number
  frozen_count?: number
  ended_count?: number
}

export interface RawContestPageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
  summary?: RawContestListSummary
}

export interface RawContestAnnouncement {
  id: string | number
  title: string
  content?: string
  created_at: string
}

export interface RawAWDRoundItem {
  id: string | number
  contest_id: string | number
  round_number: number
  status: AWDRoundData['status']
  started_at?: string | null
  ended_at?: string | null
  attack_score: number
  defense_score: number
  created_at: string
  updated_at: string
}

export interface RawAWDTeamServiceItem {
  id: string | number
  round_id: string | number
  team_id: string | number
  team_name: string
  service_id?: string | number
  service_name?: string | null
  awd_challenge_id: string | number
  awd_challenge_title?: string | null
  service_status: AWDTeamServiceData['service_status']
  checker_type?: string | null
  check_result?: Record<string, unknown> | null
  attack_received: number
  sla_score?: number | null
  defense_score: number
  attack_score: number
  updated_at: string
}

export interface RawAWDAttackLogItem {
  id: string | number
  round_id: string | number
  attacker_team_id: string | number
  attacker_team: string
  victim_team_id: string | number
  victim_team: string
  service_id?: string | number
  awd_challenge_id: string | number
  attack_type: AWDAttackLogData['attack_type']
  source?: AWDAttackLogData['source']
  submitted_flag?: string
  is_success: boolean
  score_gained: number
  created_at: string
}

export interface RawAWDRoundSummaryItem {
  team_id: string | number
  team_name: string
  service_up_count: number
  service_down_count: number
  service_compromised_count: number
  sla_score?: number | null
  defense_score: number
  attack_score: number
  successful_attack_count: number
  successful_breach_count: number
  unique_attackers_against: number
  total_score: number
}

export interface RawAWDRoundMetricsData {
  total_service_count: number
  service_up_count: number
  service_down_count: number
  service_compromised_count: number
  attacked_service_count: number
  defense_success_count: number
  total_attack_count: number
  successful_attack_count: number
  failed_attack_count: number
  scheduler_check_count: number
  manual_current_round_check_count: number
  manual_selected_round_check_count: number
  manual_service_check_count: number
  submission_attack_count: number
  manual_attack_log_count: number
  legacy_attack_log_count: number
}

export interface RawAWDRoundSummaryData {
  round: RawAWDRoundItem
  metrics?: RawAWDRoundMetricsData
  items: RawAWDRoundSummaryItem[]
}

export interface RawAWDTrafficTopTeamItem {
  team_id: string | number
  team_name: string
  request_count: number
  error_count: number
}

export interface RawAWDTrafficTopChallengeItem {
  awd_challenge_id: string | number
  awd_challenge_title?: string
  request_count: number
  error_count: number
}

export interface RawAWDTrafficTopPathItem {
  path: string
  request_count: number
  error_count: number
  last_status_code?: number
}

export interface RawAWDTrafficTrendBucketItem {
  bucket_start_at: string
  bucket_end_at?: string | null
  request_count: number
  error_count: number
}

export interface RawAWDTrafficSummaryData {
  round?: RawAWDRoundItem
  contest_id: string | number
  round_id: string | number
  total_request_count: number
  active_attacker_team_count: number
  victim_team_count: number
  unique_path_count: number
  error_request_count: number
  latest_event_at?: string
  top_attackers: RawAWDTrafficTopTeamItem[]
  top_victims: RawAWDTrafficTopTeamItem[]
  top_challenges: RawAWDTrafficTopChallengeItem[]
  top_paths?: RawAWDTrafficTopPathItem[]
  top_error_paths: RawAWDTrafficTopPathItem[]
  trend_buckets: RawAWDTrafficTrendBucketItem[]
}

export interface RawAWDTrafficEventItem {
  id: string | number
  contest_id: string | number
  round_id: string | number
  attacker_team_id: string | number
  attacker_team_name?: string
  victim_team_id: string | number
  victim_team_name?: string
  service_id?: string | number
  awd_challenge_id: string | number
  awd_challenge_title?: string
  method: string
  path: string
  status_code: number
  status_group?: AWDTrafficStatusGroup
  is_error?: boolean
  source: string
  request_id?: string
  occurred_at: string
}

export interface RawAWDCheckerRunData {
  round: RawAWDRoundItem
  services: RawAWDTeamServiceItem[]
}

export interface RawAWDReadinessItemData {
  awd_challenge_id: string | number
  title: string
  checker_type?: AWDCheckerType | null
  validation_state: AWDReadinessItemData['validation_state']
  last_preview_at?: string | null
  last_access_url?: string | null
  blocking_reason: AWDReadinessBlockingReason
}

export interface RawAWDReadinessData {
  contest_id: string | number
  ready: boolean
  total_challenges: number
  passed_challenges: number
  pending_challenges: number
  failed_challenges: number
  stale_challenges: number
  missing_checker_challenges: number
  blocking_count: number
  blocking_actions: AWDReadinessAction[]
  global_blocking_reasons: AWDReadinessGlobalReason[]
  items: RawAWDReadinessItemData[]
}

export interface RawAWDCheckerPreviewContext {
  access_url: string
  preview_flag: string
  round_number: number
  team_id: string | number
  awd_challenge_id: string | number
}

export interface RawAWDCheckerPreviewData {
  checker_type?: string | null
  service_status: AWDCheckerPreviewData['service_status']
  check_result?: Record<string, unknown> | null
  preview_context?: RawAWDCheckerPreviewContext | null
  preview_token?: string | null
}

export interface RawAdminContestTeamItem {
  id: string | number
  contest_id: string | number
  name: string
  captain_id: string | number
  invite_code?: string
  max_members: number
  member_count: number
  created_at: string
}

export interface RawAdminContestChallengeItem {
  id: string | number
  contest_id: string | number
  challenge_id: string | number
  title?: string
  category?: AdminContestChallengeRelationData['category']
  difficulty?: AdminContestChallengeRelationData['difficulty']
  points: number
  order: number
  is_visible: boolean
  created_at: string
}

export interface RawAdminContestAWDServiceItem {
  id: string | number
  contest_id: string | number
  awd_challenge_id: string | number
  title?: string | null
  category?: AdminContestAWDServiceData['category'] | null
  difficulty?: AdminContestAWDServiceData['difficulty'] | null
  display_name: string
  order: number
  is_visible: boolean
  score_config?: Record<string, unknown> | null
  runtime_config?: Record<string, unknown> | null
  validation_state?: string | null
  last_preview_at?: string | null
  last_preview_result?: RawAWDCheckerPreviewData | null
  created_at: string
  updated_at: string
}

export interface RawAdminContestAWDInstanceTeamItem {
  team_id: string | number
  team_name: string
  captain_id: string | number
}

export interface RawAdminContestAWDInstanceServiceItem {
  service_id: string | number
  awd_challenge_id: string | number
  display_name: string
  is_visible: boolean
}

export interface RawAdminContestAWDInstanceItem {
  team_id: string | number
  service_id: string | number
  instance?: RawInstanceData | null
}

export interface RawAdminContestAWDInstancePrewarmItem {
  team_id: string | number
  service_id: string | number
  outcome: 'started' | 'reused' | 'failed'
  instance?: RawInstanceData | null
  error_message?: string | null
}

export interface RawAdminContestAWDInstancePrewarmSummary {
  total: number
  started: number
  reused: number
  failed: number
}

export interface RawAdminContestAWDInstanceOrchestration {
  contest_id: string | number
  teams: RawAdminContestAWDInstanceTeamItem[]
  services: RawAdminContestAWDInstanceServiceItem[]
  instances: RawAdminContestAWDInstanceItem[]
}

export interface RawAdminContestAWDInstancePrewarm {
  contest_id: string | number
  results: RawAdminContestAWDInstancePrewarmItem[]
  summary: RawAdminContestAWDInstancePrewarmSummary
}

/* ── Normalize 函数 ── */

export function normalizeContestStatus(status: RawContestItem['status']): AdminContestStatus {
  if (status === 'registration') {
    return 'registering'
  }
  return status
}

export function serializeContestStatus(
  status?: AdminContestStatus
): RawContestItem['status'] | undefined {
  if (!status) {
    return undefined
  }
  if (status === 'registering') {
    return 'registration'
  }
  return status
}

export function serializeContestStatuses(statuses?: AdminContestStatus[]): string | undefined {
  if (!statuses?.length) {
    return undefined
  }
  const serialized = statuses
    .map((status) => serializeContestStatus(status))
    .filter((status): status is RawContestItem['status'] => Boolean(status))
  if (serialized.length === 0) {
    return undefined
  }
  return serialized.join(',')
}

export function normalizeContest(item: RawContestItem): ContestDetailData {
  return {
    id: String(item.id),
    title: item.title,
    description: item.description,
    mode: item.mode,
    status: normalizeContestStatus(item.status),
    starts_at: item.start_time,
    ends_at: item.end_time,
    scoreboard_frozen: Boolean(item.freeze_time),
  }
}

export function normalizeContestSummary(
  summary?: RawContestListSummary
): ContestPageData<ContestDetailData>['summary'] {
  if (!summary) {
    return undefined
  }
  return {
    draft_count: summary.draft_count ?? 0,
    registering_count: summary.registering_count ?? 0,
    running_count: summary.running_count ?? 0,
    frozen_count: summary.frozen_count ?? 0,
    ended_count: summary.ended_count ?? 0,
  }
}

export function normalizeContestAnnouncement(item: RawContestAnnouncement): ContestAnnouncement {
  return {
    id: String(item.id),
    title: item.title,
    content: item.content,
    created_at: item.created_at,
  }
}

export function normalizeAWDRound(item: RawAWDRoundItem): AWDRoundData {
  return {
    id: String(item.id),
    contest_id: String(item.contest_id),
    round_number: item.round_number,
    status: item.status,
    started_at: item.started_at || undefined,
    ended_at: item.ended_at || undefined,
    attack_score: item.attack_score,
    defense_score: item.defense_score,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

export function normalizeAWDCheckerType(value: unknown): AWDCheckerType | undefined {
  switch (value) {
    case 'legacy_probe':
    case 'http_standard':
    case 'tcp_standard':
    case 'script_checker':
      return value
    default:
      return undefined
  }
}

export function normalizeAWDTeamService(item: RawAWDTeamServiceItem): AWDTeamServiceData {
  return {
    id: String(item.id),
    round_id: String(item.round_id),
    team_id: String(item.team_id),
    team_name: item.team_name,
    service_id: item.service_id == null ? undefined : String(item.service_id),
    service_name: item.service_name || undefined,
    awd_challenge_id: String(item.awd_challenge_id),
    awd_challenge_title: item.awd_challenge_title || undefined,
    service_status: item.service_status,
    checker_type: normalizeAWDCheckerType(item.checker_type),
    check_result: item.check_result || {},
    attack_received: item.attack_received,
    sla_score: typeof item.sla_score === 'number' ? item.sla_score : 0,
    defense_score: item.defense_score,
    attack_score: item.attack_score,
    updated_at: item.updated_at,
  }
}

export function normalizeAWDAttackLog(item: RawAWDAttackLogItem): AWDAttackLogData {
  return {
    id: String(item.id),
    round_id: String(item.round_id),
    attacker_team_id: String(item.attacker_team_id),
    attacker_team: item.attacker_team,
    victim_team_id: String(item.victim_team_id),
    victim_team: item.victim_team,
    service_id: item.service_id == null ? undefined : String(item.service_id),
    awd_challenge_id: String(item.awd_challenge_id),
    attack_type: item.attack_type,
    source: item.source || 'legacy',
    submitted_flag: item.submitted_flag || undefined,
    is_success: item.is_success,
    score_gained: item.score_gained,
    created_at: item.created_at,
  }
}

export function normalizeAWDRoundSummaryItem(item: RawAWDRoundSummaryItem): AWDRoundSummaryItemData {
  return {
    team_id: String(item.team_id),
    team_name: item.team_name,
    service_up_count: item.service_up_count,
    service_down_count: item.service_down_count,
    service_compromised_count: item.service_compromised_count,
    sla_score: typeof item.sla_score === 'number' ? item.sla_score : 0,
    defense_score: item.defense_score,
    attack_score: item.attack_score,
    successful_attack_count: item.successful_attack_count,
    successful_breach_count: item.successful_breach_count,
    unique_attackers_against: item.unique_attackers_against,
    total_score: item.total_score,
  }
}

export function normalizeNumberOrZero(value: unknown): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : 0
}

export function normalizeAWDRoundMetrics(item: RawAWDRoundMetricsData): AWDRoundMetricsData {
  return {
    total_service_count: normalizeNumberOrZero(item.total_service_count),
    service_up_count: normalizeNumberOrZero(item.service_up_count),
    service_down_count: normalizeNumberOrZero(item.service_down_count),
    service_compromised_count: normalizeNumberOrZero(item.service_compromised_count),
    attacked_service_count: normalizeNumberOrZero(item.attacked_service_count),
    defense_success_count: normalizeNumberOrZero(item.defense_success_count),
    total_attack_count: normalizeNumberOrZero(item.total_attack_count),
    successful_attack_count: normalizeNumberOrZero(item.successful_attack_count),
    failed_attack_count: normalizeNumberOrZero(item.failed_attack_count),
    scheduler_check_count: normalizeNumberOrZero(item.scheduler_check_count),
    manual_current_round_check_count: normalizeNumberOrZero(item.manual_current_round_check_count),
    manual_selected_round_check_count: normalizeNumberOrZero(
      item.manual_selected_round_check_count
    ),
    manual_service_check_count: normalizeNumberOrZero(item.manual_service_check_count),
    submission_attack_count: normalizeNumberOrZero(item.submission_attack_count),
    manual_attack_log_count: normalizeNumberOrZero(item.manual_attack_log_count),
    legacy_attack_log_count: normalizeNumberOrZero(item.legacy_attack_log_count),
  }
}

export function normalizeAWDRoundSummary(item: RawAWDRoundSummaryData): AWDRoundSummaryData {
  return {
    round: normalizeAWDRound(item.round),
    metrics: item.metrics ? normalizeAWDRoundMetrics(item.metrics) : undefined,
    items: item.items.map(normalizeAWDRoundSummaryItem),
  }
}

export function normalizeAWDTrafficTopTeam(
  item: RawAWDTrafficTopTeamItem
): AWDTrafficTopTeamData {
  return {
    team_id: String(item.team_id),
    team_name: item.team_name,
    request_count: item.request_count,
    error_count: item.error_count,
  }
}

export function normalizeAWDTrafficTopChallenge(
  item: RawAWDTrafficTopChallengeItem
): AWDTrafficTopChallengeData {
  return {
    awd_challenge_id: String(item.awd_challenge_id),
    awd_challenge_title: item.awd_challenge_title,
    request_count: item.request_count,
    error_count: item.error_count,
  }
}

export function normalizeAWDTrafficTopPath(
  item: RawAWDTrafficTopPathItem
): AWDTrafficTopPathData {
  return {
    path: item.path,
    request_count: item.request_count,
    error_count: item.error_count,
    last_status_code: item.last_status_code ?? 0,
  }
}

export function normalizeAWDTrafficTrendBucket(
  item: RawAWDTrafficTrendBucketItem
): AWDTrafficTrendBucketData {
  return {
    bucket_start_at: item.bucket_start_at,
    bucket_end_at: item.bucket_end_at || undefined,
    request_count: item.request_count,
    error_count: item.error_count,
  }
}

export function normalizeAWDTrafficSummary(
  item: RawAWDTrafficSummaryData
): AWDTrafficSummaryData {
  return {
    round: item.round ? normalizeAWDRound(item.round) : undefined,
    contest_id: String(item.contest_id),
    round_id: String(item.round_id),
    total_request_count: item.total_request_count,
    active_attacker_team_count: item.active_attacker_team_count,
    victim_team_count: item.victim_team_count,
    unique_path_count: item.unique_path_count,
    error_request_count: item.error_request_count,
    latest_event_at: item.latest_event_at,
    top_attackers: item.top_attackers.map(normalizeAWDTrafficTopTeam),
    top_victims: item.top_victims.map(normalizeAWDTrafficTopTeam),
    top_challenges: item.top_challenges.map(normalizeAWDTrafficTopChallenge),
    top_paths: (item.top_paths || []).map(normalizeAWDTrafficTopPath),
    top_error_paths: item.top_error_paths.map(normalizeAWDTrafficTopPath),
    trend_buckets: item.trend_buckets.map(normalizeAWDTrafficTrendBucket),
  }
}

export function classifyAWDTrafficStatusGroup(statusCode: number): AWDTrafficStatusGroup {
  if (statusCode >= 500) {
    return 'server_error'
  }
  if (statusCode >= 400) {
    return 'client_error'
  }
  if (statusCode >= 300) {
    return 'redirect'
  }
  return 'success'
}

export function normalizeAWDTrafficEvent(item: RawAWDTrafficEventItem): AWDTrafficEventData {
  const statusGroup = item.status_group || classifyAWDTrafficStatusGroup(item.status_code)
  return {
    id: String(item.id),
    contest_id: String(item.contest_id),
    round_id: String(item.round_id),
    attacker_team_id: String(item.attacker_team_id),
    attacker_team_name: item.attacker_team_name,
    victim_team_id: String(item.victim_team_id),
    victim_team_name: item.victim_team_name,
    service_id: item.service_id == null ? undefined : String(item.service_id),
    awd_challenge_id: String(item.awd_challenge_id),
    awd_challenge_title: item.awd_challenge_title,
    method: item.method,
    path: item.path,
    status_code: item.status_code,
    status_group: statusGroup,
    is_error:
      item.is_error ?? (statusGroup === 'client_error' || statusGroup === 'server_error'),
    source: item.source,
    request_id: item.request_id,
    occurred_at: item.occurred_at,
  }
}

export function normalizeContestScoreboard(item: ContestScoreboardData): ContestScoreboardData {
  return {
    contest: {
      id: String(item.contest.id),
      title: item.contest.title,
      status: item.contest.status,
      started_at: item.contest.started_at,
      ends_at: item.contest.ends_at,
    },
    scoreboard: {
      ...item.scoreboard,
      list: item.scoreboard.list.map((row) => ({
        ...row,
        team_id: String(row.team_id),
      })),
    },
    frozen: item.frozen,
  }
}

export function normalizeAWDCheckerRun(item: RawAWDCheckerRunData): AWDCheckerRunData {
  return {
    round: normalizeAWDRound(item.round),
    services: item.services.map(normalizeAWDTeamService),
  }
}

export function normalizeAWDCheckerPreview(
  item: RawAWDCheckerPreviewData
): AWDCheckerPreviewData {
  return {
    checker_type: normalizeAWDCheckerType(item.checker_type),
    service_status: item.service_status,
    check_result: item.check_result || {},
    preview_context: {
      access_url: item.preview_context?.access_url || '',
      preview_flag: item.preview_context?.preview_flag || '',
      round_number: item.preview_context?.round_number ?? 0,
      team_id: String(item.preview_context?.team_id ?? 0),
      awd_challenge_id: String(item.preview_context?.awd_challenge_id ?? 0),
    },
    preview_token: item.preview_token || undefined,
  }
}

export function normalizeAWDReadinessItem(
  item: RawAWDReadinessItemData
): AWDReadinessItemData {
  return {
    awd_challenge_id: String(item.awd_challenge_id),
    title: item.title,
    checker_type: item.checker_type || undefined,
    validation_state: item.validation_state,
    last_preview_at: item.last_preview_at || undefined,
    last_access_url: item.last_access_url || undefined,
    blocking_reason: item.blocking_reason,
  }
}

export function normalizeAWDReadiness(item: RawAWDReadinessData): AWDReadinessData {
  return {
    contest_id: String(item.contest_id),
    ready: item.ready,
    total_challenges: item.total_challenges,
    passed_challenges: item.passed_challenges,
    pending_challenges: item.pending_challenges,
    failed_challenges: item.failed_challenges,
    stale_challenges: item.stale_challenges,
    missing_checker_challenges: item.missing_checker_challenges,
    blocking_count: item.blocking_count,
    blocking_actions: item.blocking_actions,
    global_blocking_reasons: item.global_blocking_reasons,
    items: item.items.map(normalizeAWDReadinessItem),
  }
}

export function normalizeAdminContestTeam(
  item: RawAdminContestTeamItem
): AdminContestTeamData {
  return {
    id: String(item.id),
    contest_id: String(item.contest_id),
    name: item.name,
    captain_id: String(item.captain_id),
    invite_code: item.invite_code,
    max_members: item.max_members,
    member_count: item.member_count,
    created_at: item.created_at,
  }
}

export function normalizeAdminContestChallenge(
  item: RawAdminContestChallengeItem
): AdminContestChallengeRelationData {
  return {
    id: String(item.id),
    contest_id: String(item.contest_id),
    challenge_id: String(item.challenge_id),
    title: item.title,
    category: item.category,
    difficulty: item.difficulty,
    points: item.points,
    order: item.order,
    is_visible: item.is_visible,
    created_at: item.created_at,
  }
}

export function normalizeContestAWDServiceCheckerConfig(
  runtimeConfig?: Record<string, unknown> | null
): Record<string, unknown> {
  if (!runtimeConfig) {
    return {}
  }
  const rawString = runtimeConfig.checker_config_raw
  if (typeof rawString === 'string' && rawString.trim()) {
    try {
      const parsed = JSON.parse(rawString)
      return parsed && typeof parsed === 'object' ? (parsed as Record<string, unknown>) : {}
    } catch {
      return {}
    }
  }
  const rawConfig = runtimeConfig.checker_config
  if (rawConfig && typeof rawConfig === 'object') {
    return rawConfig as Record<string, unknown>
  }
  return {}
}

export function normalizeContestAWDServiceScore(value: unknown): number | undefined {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  return undefined
}

export function normalizeContestAWDServiceValidationState(
  value: unknown
): AdminContestAWDServiceData['validation_state'] {
  if (value === 'pending' || value === 'failed' || value === 'stale' || value === 'passed') {
    return value
  }
  return undefined
}

export function normalizeAdminContestAWDService(
  item: RawAdminContestAWDServiceItem
): AdminContestAWDServiceData {
  const runtimeConfig = { ...(item.runtime_config || {}) }
  delete runtimeConfig.challenge_id
  delete runtimeConfig.awd_challenge_id
  const scoreConfig = item.score_config || {}
  return {
    id: String(item.id),
    contest_id: String(item.contest_id),
    awd_challenge_id: String(item.awd_challenge_id),
    title: item.title || undefined,
    category: item.category || undefined,
    difficulty: item.difficulty || undefined,
    display_name: item.display_name,
    order: item.order,
    is_visible: item.is_visible,
    score_config: scoreConfig,
    runtime_config: runtimeConfig,
    checker_type: normalizeAWDCheckerType(runtimeConfig.checker_type),
    checker_config: normalizeContestAWDServiceCheckerConfig(runtimeConfig),
    sla_score: normalizeContestAWDServiceScore(scoreConfig.awd_sla_score),
    defense_score: normalizeContestAWDServiceScore(scoreConfig.awd_defense_score),
    validation_state: normalizeContestAWDServiceValidationState(item.validation_state),
    last_preview_at: item.last_preview_at || undefined,
    last_preview_result: item.last_preview_result
      ? normalizeAWDCheckerPreview(item.last_preview_result)
      : undefined,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

export function normalizeAdminContestAWDInstanceTeam(
  item: RawAdminContestAWDInstanceTeamItem
): AdminContestAWDInstanceTeamData {
  return {
    team_id: String(item.team_id),
    team_name: item.team_name,
    captain_id: String(item.captain_id),
  }
}

export function normalizeAdminContestAWDInstanceService(
  item: RawAdminContestAWDInstanceServiceItem
): AdminContestAWDInstanceServiceData {
  return {
    service_id: String(item.service_id),
    awd_challenge_id: String(item.awd_challenge_id),
    display_name: item.display_name,
    is_visible: item.is_visible,
  }
}

export function normalizeAdminContestAWDInstanceItem(
  item: RawAdminContestAWDInstanceItem
): AdminContestAWDInstanceItemData {
  return {
    team_id: String(item.team_id),
    service_id: String(item.service_id),
    instance: item.instance ? normalizeInstanceData(item.instance) : undefined,
  }
}

export function normalizeAdminContestAWDInstancePrewarmItem(
  item: RawAdminContestAWDInstancePrewarmItem
): AdminContestAWDInstancePrewarmItemData {
  return {
    team_id: String(item.team_id),
    service_id: String(item.service_id),
    outcome: item.outcome,
    instance: item.instance ? normalizeInstanceData(item.instance) : undefined,
    error_message: item.error_message || undefined,
  }
}

export function normalizeAdminContestAWDInstanceOrchestration(
  item: RawAdminContestAWDInstanceOrchestration
): AdminContestAWDInstanceOrchestrationData {
  return {
    contest_id: String(item.contest_id),
    teams: item.teams.map(normalizeAdminContestAWDInstanceTeam),
    services: item.services.map(normalizeAdminContestAWDInstanceService),
    instances: item.instances.map(normalizeAdminContestAWDInstanceItem),
  }
}

export function normalizeAdminContestAWDInstancePrewarm(
  item: RawAdminContestAWDInstancePrewarm
): AdminContestAWDInstancePrewarmData {
  return {
    contest_id: String(item.contest_id),
    results: item.results.map(normalizeAdminContestAWDInstancePrewarmItem),
    summary: {
      total: item.summary?.total ?? 0,
      started: item.summary?.started ?? 0,
      reused: item.summary?.reused ?? 0,
      failed: item.summary?.failed ?? 0,
    },
  }
}

export function serializeContestPayload(
  data: AdminContestCreatePayload | AdminContestUpdatePayload
) {
  return {
    title: data.title,
    description: data.description,
    mode: data.mode,
    start_time: data.starts_at,
    end_time: data.ends_at,
    status: 'status' in data ? serializeContestStatus(data.status) : undefined,
    force_override: 'force_override' in data ? data.force_override : undefined,
    override_reason: 'override_reason' in data ? data.override_reason : undefined,
  }
}
