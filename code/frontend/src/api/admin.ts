import { ApiError, request } from './request'

import type {
  AWDAttackLogData,
  AWDCheckerPreviewData,
  AWDCheckerType,
  AWDCheckerRunData,
  AWDReadinessAction,
  AWDReadinessBlockingReason,
  AWDReadinessData,
  AWDReadinessGlobalReason,
  AWDReadinessItemData,
  AWDRoundData,
  AWDRoundMetricsData,
  AWDRoundSummaryData,
  AWDRoundSummaryItemData,
  AWDTrafficEventData,
  AWDTrafficEventPageData,
  AWDTrafficStatusGroup,
  AWDTrafficSummaryData,
  AWDTrafficTopChallengeData,
  AWDTrafficTopPathData,
  AWDTrafficTopTeamData,
  AWDTrafficTrendBucketData,
  AWDTeamServiceData,
  AdminContestChallengeData,
  AdminContestTeamData,
  AdminChallengeHint,
  AdminChallengeImportCommitData,
  AdminChallengeImportPreview,
  AdminChallengeListItem,
  AdminChallengePublishRequestData,
  AdminNotificationPublishPayload,
  AdminNotificationPublishResult,
  AdminChallengeWriteupData,
  AdminCheatDetectionData,
  AdminDashboardData,
  AdminImageListItem,
  AdminUserImportData,
  AdminUserListItem,
  AdminUserUpsertData,
  AuditLogItem,
  ChallengeTopologyData,
  ContestMode,
  ContestDetailData,
  ContestScoreboardData,
  ContestStatus,
  EnvironmentTemplateData,
  PageResult,
  TopologyLinkData,
  TopologyNetworkData,
  TopologyNodeData,
  TopologyTrafficPolicyData,
  WriteupVisibility,
  ReportExportData,
} from './contracts'
import type { UserRole } from '@/utils/constants'

type AdminContestStatus = Extract<
  ContestStatus,
  'draft' | 'registering' | 'running' | 'frozen' | 'ended'
>
type AdminContestMode = Extract<ContestMode, 'jeopardy' | 'awd'>
type UserStatus = 'active' | 'inactive' | 'locked' | 'banned'

interface UserListParams {
  page?: number
  page_size?: number
  keyword?: string
  student_no?: string
  teacher_no?: string
  role?: UserRole
  status?: UserStatus
  class_name?: string
}

export interface AdminUserCreatePayload {
  username: string
  name?: string
  password: string
  email?: string
  student_no?: string
  teacher_no?: string
  class_name?: string
  role: UserRole
  status?: UserStatus
}

export interface AdminUserUpdatePayload {
  name?: string
  password?: string
  email?: string
  student_no?: string
  teacher_no?: string
  class_name?: string
  role?: UserRole
  status?: UserStatus
}

export interface AdminContestChallengeCreatePayload {
  challenge_id: number
  points: number
  order?: number
  is_visible?: boolean
  awd_checker_type?: AWDCheckerType
  awd_checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}

export interface AdminContestChallengeUpdatePayload {
  points?: number
  order?: number
  is_visible?: boolean
  awd_checker_type?: AWDCheckerType
  awd_checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}

export interface AdminAWDCheckerPreviewPayload {
  challenge_id: number
  checker_type: AWDCheckerType
  checker_config: Record<string, unknown>
  access_url: string
  preview_flag?: string
}

export async function exportContestArchive(
  contestId: string,
  data?: { format?: 'json' }
): Promise<ReportExportData> {
  const payload = await request<ReportExportData & { report_id: string | number }>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/export`,
    data,
  })

  return {
    ...payload,
    report_id: String(payload.report_id),
  }
}

interface RawContestItem {
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

interface RawAWDRoundItem {
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

interface RawAWDTeamServiceItem {
  id: string | number
  round_id: string | number
  team_id: string | number
  team_name: string
  challenge_id: string | number
  service_status: AWDTeamServiceData['service_status']
  checker_type?: string | null
  check_result?: Record<string, unknown> | null
  attack_received: number
  sla_score?: number | null
  defense_score: number
  attack_score: number
  updated_at: string
}

interface RawAWDAttackLogItem {
  id: string | number
  round_id: string | number
  attacker_team_id: string | number
  attacker_team: string
  victim_team_id: string | number
  victim_team: string
  challenge_id: string | number
  attack_type: AWDAttackLogData['attack_type']
  source?: AWDAttackLogData['source']
  submitted_flag?: string
  is_success: boolean
  score_gained: number
  created_at: string
}

interface RawAdminNotificationPublishResult {
  batch_id: string | number
  recipient_count: number
}

interface RawAWDRoundSummaryItem {
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

interface RawAWDRoundMetricsData {
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

interface RawAWDRoundSummaryData {
  round: RawAWDRoundItem
  metrics?: RawAWDRoundMetricsData
  items: RawAWDRoundSummaryItem[]
}

interface RawAWDTrafficTopTeamItem {
  team_id: string | number
  team_name: string
  request_count: number
  error_count: number
}

interface RawAWDTrafficTopChallengeItem {
  challenge_id: string | number
  challenge_title?: string
  request_count: number
  error_count: number
}

interface RawAWDTrafficTopPathItem {
  path: string
  request_count: number
  error_count: number
  last_status_code?: number
}

interface RawAWDTrafficTrendBucketItem {
  bucket_start_at: string
  bucket_end_at?: string | null
  request_count: number
  error_count: number
}

interface RawAWDTrafficSummaryData {
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

interface RawAWDTrafficEventItem {
  id: string | number
  contest_id: string | number
  round_id: string | number
  attacker_team_id: string | number
  attacker_team_name?: string
  victim_team_id: string | number
  victim_team_name?: string
  challenge_id: string | number
  challenge_title?: string
  method: string
  path: string
  status_code: number
  status_group?: AWDTrafficStatusGroup
  is_error?: boolean
  source: string
  request_id?: string
  occurred_at: string
}

interface RawAWDCheckerRunData {
  round: RawAWDRoundItem
  services: RawAWDTeamServiceItem[]
}

interface RawAWDReadinessItemData {
  challenge_id: string | number
  title: string
  checker_type?: AWDCheckerType | null
  validation_state: AWDReadinessItemData['validation_state']
  last_preview_at?: string | null
  last_access_url?: string | null
  blocking_reason: AWDReadinessBlockingReason
}

interface RawAWDReadinessData {
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

interface RawAWDCheckerPreviewContext {
  access_url: string
  preview_flag: string
  round_number: number
  team_id: string | number
  challenge_id: string | number
}

interface RawAWDCheckerPreviewData {
  checker_type?: string | null
  service_status: AWDCheckerPreviewData['service_status']
  check_result?: Record<string, unknown> | null
  preview_context?: RawAWDCheckerPreviewContext | null
  preview_token?: string | null
}

interface RawAdminContestTeamItem {
  id: string | number
  contest_id: string | number
  name: string
  captain_id: string | number
  invite_code?: string
  max_members: number
  member_count: number
  created_at: string
}

interface RawAdminContestChallengeItem {
  id: string | number
  contest_id: string | number
  challenge_id: string | number
  title?: string
  category?: AdminContestChallengeData['category']
  difficulty?: AdminContestChallengeData['difficulty']
  points: number
  order: number
  is_visible: boolean
  awd_checker_type?: string | null
  awd_checker_config?: Record<string, unknown> | null
  awd_sla_score?: number | null
  awd_defense_score?: number | null
  awd_checker_validation_state?: AdminContestChallengeData['awd_checker_validation_state'] | null
  awd_checker_last_preview_at?: string | null
  awd_checker_last_preview_result?: RawAWDCheckerPreviewData | null
  created_at: string
}

interface RawAdminChallengeItem {
  id: string | number
  title: string
  description?: string
  category: AdminChallengeListItem['category']
  difficulty: AdminChallengeListItem['difficulty']
  points: number
  instance_sharing?: AdminChallengeListItem['instance_sharing']
  created_by?: string | number | null
  image_id?: string | number | null
  attachment_url?: string
  hints?: Array<{
    id: string | number
    level: number
    title?: string
    content: string
  }>
  status: 'draft' | 'published' | 'archived'
  created_at: string
  updated_at: string
}

interface RawAdminChallengePublishRequestData {
  id: string | number
  challenge_id: string | number
  status: AdminChallengePublishRequestData['status'] | 'pending' | 'passed'
  requested_by?: string | number | null
  request_source?: string | null
  active?: boolean
  failure_summary?: string | null
  started_at?: string | null
  finished_at?: string | null
  published_at?: string | null
  result?: {
    challenge_id: string | number
    precheck: {
      passed: boolean
      started_at: string
      ended_at: string
      steps: Array<{
        name: string
        passed: boolean
        message: string
      }>
    }
    runtime: {
      passed: boolean
      started_at: string
      ended_at: string
      access_url?: string
      container_count: number
      network_count: number
      steps: Array<{
        name: string
        passed: boolean
        message: string
      }>
    }
  } | null
  created_at: string
  updated_at: string
}

interface RawChallengeImportPreview {
  id: string | number
  file_name: string
  slug: string
  title: string
  description: string
  category: AdminChallengeListItem['category']
  difficulty: AdminChallengeListItem['difficulty']
  points: number
  attachments?: Array<{
    name: string
    path: string
  }>
  hints?: Array<{
    id?: string | number
    level: number
    title?: string
    content: string
  }>
  flag: {
    type: 'static' | 'dynamic' | 'regex' | 'manual_review'
    prefix?: string
  }
  runtime: {
    type?: string
    image_ref?: string
  }
  extensions: {
    topology: {
      source?: string
      enabled: boolean
    }
  }
  warnings?: string[]
  created_at: string
}

interface RawTopologyNetworkData {
  key: string
  name: string
  cidr?: string
  internal?: boolean
}

interface RawTopologyNodeResourcesData {
  cpu_quota?: number
  memory_mb?: number
  pids_limit?: number
}

interface RawTopologyNodeData {
  key: string
  name: string
  image_id?: string | number | null
  service_port?: number
  inject_flag?: boolean
  tier?: TopologyNodeData['tier']
  network_keys?: string[]
  env?: Record<string, string>
  resources?: RawTopologyNodeResourcesData
}

interface RawTopologyLinkData {
  from_node_key: string
  to_node_key: string
}

interface RawTopologyTrafficPolicyData {
  source_node_key: string
  target_node_key: string
  action: TopologyTrafficPolicyData['action']
  protocol?: TopologyTrafficPolicyData['protocol']
  ports?: number[]
}

interface RawChallengeTopologyData {
  id: string | number
  challenge_id: string | number
  template_id?: string | number | null
  entry_node_key: string
  networks?: RawTopologyNetworkData[]
  nodes: RawTopologyNodeData[]
  links?: RawTopologyLinkData[]
  policies?: RawTopologyTrafficPolicyData[]
  created_at: string
  updated_at: string
}

interface RawEnvironmentTemplateData {
  id: string | number
  name: string
  description: string
  entry_node_key: string
  networks?: RawTopologyNetworkData[]
  nodes: RawTopologyNodeData[]
  links?: RawTopologyLinkData[]
  policies?: RawTopologyTrafficPolicyData[]
  usage_count: number
  created_at: string
  updated_at: string
}

interface RawChallengeFlagConfig {
  flag_type: 'static' | 'dynamic' | 'regex' | 'manual_review'
  flag_regex?: string
  flag_prefix?: string
  configured: boolean
}

interface RawAdminChallengeWriteupData {
  id: string | number
  challenge_id: string | number
  title: string
  content: string
  visibility: WriteupVisibility
  created_by?: string | number | null
  is_recommended?: boolean
  recommended_at?: string | null
  recommended_by?: string | number | null
  created_at: string
  updated_at: string
}

interface RawImageItem {
  id: string | number
  name: string
  tag: string
  description?: string
  size?: number
  status: AdminImageListItem['status']
  created_at: string
  updated_at: string
}

interface RawAdminUser {
  id: string | number
  username: string
  name?: string | null
  email?: string | null
  student_no?: string | null
  teacher_no?: string | null
  class_name?: string | null
  status: UserStatus
  roles: UserRole[]
  created_at: string
}

interface RawCheatDetectionData {
  generated_at: string
  summary: {
    submit_burst_users: number
    shared_ip_groups: number
    affected_users: number
  }
  suspects: Array<{
    user_id: string | number
    username: string
    submit_count: number
    last_seen_at: string
    reason: string
  }>
  shared_ips: Array<{
    ip: string
    user_count: number
    usernames: string[]
  }>
}

interface ContestListParams {
  page?: number
  page_size?: number
  status?: AdminContestStatus
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

interface AdminRequestOptions {
  suppressErrorToast?: boolean
}

export interface AdminAWDServiceCheckPayload {
  team_id: number
  challenge_id: number
  service_status: AWDTeamServiceData['service_status']
  check_result?: Record<string, unknown>
}

export interface AdminAWDAttackLogPayload {
  attacker_team_id: number
  victim_team_id: number
  challenge_id: number
  attack_type: AWDAttackLogData['attack_type']
  submitted_flag?: string
  is_success: boolean
}

export interface AdminAWDTrafficEventsParams {
  attacker_team_id?: string
  victim_team_id?: string
  challenge_id?: string
  status_group?: AWDTrafficStatusGroup
  path_keyword?: string
  page?: number
  page_size?: number
}

function normalizeContestStatus(status: RawContestItem['status']): AdminContestStatus {
  if (status === 'registration') {
    return 'registering'
  }
  return status
}

function normalizeAdminChallengeWriteup(
  item: RawAdminChallengeWriteupData
): AdminChallengeWriteupData {
  return {
    id: String(item.id),
    challenge_id: String(item.challenge_id),
    title: item.title,
    content: item.content,
    visibility: item.visibility,
    created_by: item.created_by == null ? undefined : String(item.created_by),
    is_recommended: item.is_recommended ?? false,
    recommended_at: item.recommended_at || undefined,
    recommended_by: item.recommended_by == null ? undefined : String(item.recommended_by),
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function serializeContestStatus(status?: AdminContestStatus): RawContestItem['status'] | undefined {
  if (!status) {
    return undefined
  }
  if (status === 'registering') {
    return 'registration'
  }
  return status
}

function normalizeContest(item: RawContestItem): ContestDetailData {
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

function normalizeAWDRound(item: RawAWDRoundItem): AWDRoundData {
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

function normalizeAWDCheckerType(value: unknown): AWDCheckerType | undefined {
  switch (value) {
    case 'legacy_probe':
    case 'http_standard':
      return value
    default:
      return undefined
  }
}

function normalizeAWDTeamService(item: RawAWDTeamServiceItem): AWDTeamServiceData {
  return {
    id: String(item.id),
    round_id: String(item.round_id),
    team_id: String(item.team_id),
    team_name: item.team_name,
    challenge_id: String(item.challenge_id),
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

function normalizeAWDAttackLog(item: RawAWDAttackLogItem): AWDAttackLogData {
  return {
    id: String(item.id),
    round_id: String(item.round_id),
    attacker_team_id: String(item.attacker_team_id),
    attacker_team: item.attacker_team,
    victim_team_id: String(item.victim_team_id),
    victim_team: item.victim_team,
    challenge_id: String(item.challenge_id),
    attack_type: item.attack_type,
    source: item.source || 'legacy',
    submitted_flag: item.submitted_flag || undefined,
    is_success: item.is_success,
    score_gained: item.score_gained,
    created_at: item.created_at,
  }
}

function normalizeAWDRoundSummaryItem(item: RawAWDRoundSummaryItem): AWDRoundSummaryItemData {
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

function normalizeAWDRoundMetrics(item: RawAWDRoundMetricsData): AWDRoundMetricsData {
  return {
    total_service_count: item.total_service_count,
    service_up_count: item.service_up_count,
    service_down_count: item.service_down_count,
    service_compromised_count: item.service_compromised_count,
    attacked_service_count: item.attacked_service_count,
    defense_success_count: item.defense_success_count,
    total_attack_count: item.total_attack_count,
    successful_attack_count: item.successful_attack_count,
    failed_attack_count: item.failed_attack_count,
    scheduler_check_count: item.scheduler_check_count,
    manual_current_round_check_count: item.manual_current_round_check_count,
    manual_selected_round_check_count: item.manual_selected_round_check_count,
    manual_service_check_count: item.manual_service_check_count,
    submission_attack_count: item.submission_attack_count,
    manual_attack_log_count: item.manual_attack_log_count,
    legacy_attack_log_count: item.legacy_attack_log_count,
  }
}

function normalizeAWDRoundSummary(item: RawAWDRoundSummaryData): AWDRoundSummaryData {
  return {
    round: normalizeAWDRound(item.round),
    metrics: item.metrics ? normalizeAWDRoundMetrics(item.metrics) : undefined,
    items: item.items.map(normalizeAWDRoundSummaryItem),
  }
}

function normalizeAWDTrafficTopTeam(item: RawAWDTrafficTopTeamItem): AWDTrafficTopTeamData {
  return {
    team_id: String(item.team_id),
    team_name: item.team_name,
    request_count: item.request_count,
    error_count: item.error_count,
  }
}

function normalizeAWDTrafficTopChallenge(
  item: RawAWDTrafficTopChallengeItem
): AWDTrafficTopChallengeData {
  return {
    challenge_id: String(item.challenge_id),
    challenge_title: item.challenge_title,
    request_count: item.request_count,
    error_count: item.error_count,
  }
}

function normalizeAWDTrafficTopPath(item: RawAWDTrafficTopPathItem): AWDTrafficTopPathData {
  return {
    path: item.path,
    request_count: item.request_count,
    error_count: item.error_count,
    last_status_code: item.last_status_code ?? 0,
  }
}

function normalizeAWDTrafficTrendBucket(
  item: RawAWDTrafficTrendBucketItem
): AWDTrafficTrendBucketData {
  return {
    bucket_start_at: item.bucket_start_at,
    bucket_end_at: item.bucket_end_at || undefined,
    request_count: item.request_count,
    error_count: item.error_count,
  }
}

function normalizeAWDTrafficSummary(item: RawAWDTrafficSummaryData): AWDTrafficSummaryData {
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

function classifyAWDTrafficStatusGroup(statusCode: number): AWDTrafficStatusGroup {
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

function normalizeAWDTrafficEvent(item: RawAWDTrafficEventItem): AWDTrafficEventData {
  const statusGroup = item.status_group || classifyAWDTrafficStatusGroup(item.status_code)
  return {
    id: String(item.id),
    contest_id: String(item.contest_id),
    round_id: String(item.round_id),
    attacker_team_id: String(item.attacker_team_id),
    attacker_team_name: item.attacker_team_name,
    victim_team_id: String(item.victim_team_id),
    victim_team_name: item.victim_team_name,
    challenge_id: String(item.challenge_id),
    challenge_title: item.challenge_title,
    method: item.method,
    path: item.path,
    status_code: item.status_code,
    status_group: statusGroup,
    is_error: item.is_error ?? (statusGroup === 'client_error' || statusGroup === 'server_error'),
    source: item.source,
    request_id: item.request_id,
    occurred_at: item.occurred_at,
  }
}

function normalizeContestScoreboard(item: ContestScoreboardData): ContestScoreboardData {
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

function normalizeAWDCheckerRun(item: RawAWDCheckerRunData): AWDCheckerRunData {
  return {
    round: normalizeAWDRound(item.round),
    services: item.services.map(normalizeAWDTeamService),
  }
}

function normalizeAWDCheckerPreview(item: RawAWDCheckerPreviewData): AWDCheckerPreviewData {
  return {
    checker_type: normalizeAWDCheckerType(item.checker_type),
    service_status: item.service_status,
    check_result: item.check_result || {},
    preview_context: {
      access_url: item.preview_context?.access_url || '',
      preview_flag: item.preview_context?.preview_flag || '',
      round_number: item.preview_context?.round_number ?? 0,
      team_id: String(item.preview_context?.team_id ?? 0),
      challenge_id: String(item.preview_context?.challenge_id ?? 0),
    },
    preview_token: item.preview_token || undefined,
  }
}

function normalizeAWDReadinessItem(item: RawAWDReadinessItemData): AWDReadinessItemData {
  return {
    challenge_id: String(item.challenge_id),
    title: item.title,
    checker_type: item.checker_type || undefined,
    validation_state: item.validation_state,
    last_preview_at: item.last_preview_at || undefined,
    last_access_url: item.last_access_url || undefined,
    blocking_reason: item.blocking_reason,
  }
}

function normalizeAWDReadiness(item: RawAWDReadinessData): AWDReadinessData {
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

function normalizeAdminContestTeam(item: RawAdminContestTeamItem): AdminContestTeamData {
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

function normalizeAdminContestChallenge(item: RawAdminContestChallengeItem): AdminContestChallengeData {
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
    awd_checker_type: normalizeAWDCheckerType(item.awd_checker_type),
    awd_checker_config: item.awd_checker_config || {},
    awd_sla_score: typeof item.awd_sla_score === 'number' ? item.awd_sla_score : 0,
    awd_defense_score: typeof item.awd_defense_score === 'number' ? item.awd_defense_score : 0,
    awd_checker_validation_state: item.awd_checker_validation_state || 'pending',
    awd_checker_last_preview_at: item.awd_checker_last_preview_at || undefined,
    awd_checker_last_preview_result: item.awd_checker_last_preview_result
      ? normalizeAWDCheckerPreview(item.awd_checker_last_preview_result)
      : undefined,
    created_at: item.created_at,
  }
}

function normalizeChallenge(
  item: RawAdminChallengeItem,
  flagConfig?: RawChallengeFlagConfig
): AdminChallengeListItem {
  const hints: AdminChallengeHint[] | undefined = item.hints?.map((hint) => ({
    id: String(hint.id),
    level: hint.level,
    title: hint.title,
    content: hint.content,
  }))

  return {
    id: String(item.id),
    title: item.title,
    description: item.description,
    category: item.category,
    difficulty: item.difficulty,
    points: item.points,
    status: item.status,
    instance_sharing: item.instance_sharing ?? 'per_user',
    created_by: item.created_by == null ? undefined : String(item.created_by),
    image_id: item.image_id == null ? undefined : String(item.image_id),
    attachment_url: item.attachment_url,
    hints,
    created_at: item.created_at,
    updated_at: item.updated_at,
    flag_config: flagConfig
      ? {
          configured: flagConfig.configured,
          flag_type: flagConfig.flag_type,
          flag_regex: flagConfig.flag_regex,
          flag_prefix: flagConfig.flag_prefix,
        }
      : undefined,
  }
}

function normalizeChallengePublishRequest(
  item: RawAdminChallengePublishRequestData
): AdminChallengePublishRequestData {
  const status =
    item.status === 'pending'
      ? 'queued'
      : item.status === 'passed'
        ? 'succeeded'
        : item.status

  return {
    id: String(item.id),
    challenge_id: String(item.challenge_id),
    status,
    active: item.active ?? (status === 'queued' || status === 'running'),
    requested_by: item.requested_by == null ? undefined : String(item.requested_by),
    request_source: item.request_source ?? undefined,
    failure_summary: item.failure_summary ?? undefined,
    started_at: item.started_at ?? undefined,
    finished_at: item.finished_at ?? undefined,
    published_at: item.published_at ?? undefined,
    result:
      item.result == null
        ? undefined
        : {
            challenge_id: String(item.result.challenge_id),
            precheck: {
              passed: item.result.precheck.passed,
              started_at: item.result.precheck.started_at,
              ended_at: item.result.precheck.ended_at,
              steps: item.result.precheck.steps.map((step) => ({
                name: step.name,
                passed: step.passed,
                message: step.message,
              })),
            },
            runtime: {
              passed: item.result.runtime.passed,
              started_at: item.result.runtime.started_at,
              ended_at: item.result.runtime.ended_at,
              access_url: item.result.runtime.access_url,
              container_count: item.result.runtime.container_count,
              network_count: item.result.runtime.network_count,
              steps: item.result.runtime.steps.map((step) => ({
                name: step.name,
                passed: step.passed,
                message: step.message,
              })),
            },
          },
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function normalizeChallengeImportPreview(item: RawChallengeImportPreview): AdminChallengeImportPreview {
  return {
    id: String(item.id),
    file_name: item.file_name,
    slug: item.slug,
    title: item.title,
    description: item.description,
    category: item.category,
    difficulty: item.difficulty,
    points: item.points,
    attachments: item.attachments,
    hints: item.hints?.map((hint) => ({
      id: hint.id == null ? undefined : String(hint.id),
      level: hint.level,
      title: hint.title,
      content: hint.content,
    })),
    flag: item.flag,
    runtime: item.runtime,
    extensions: item.extensions,
    warnings: item.warnings,
    created_at: item.created_at,
  }
}

function normalizeTopologyNetwork(item: RawTopologyNetworkData): TopologyNetworkData {
  return {
    key: item.key,
    name: item.name,
    cidr: item.cidr,
    internal: item.internal,
  }
}

function normalizeTopologyNode(item: RawTopologyNodeData): TopologyNodeData {
  return {
    key: item.key,
    name: item.name,
    image_id: item.image_id == null ? undefined : String(item.image_id),
    service_port: item.service_port,
    inject_flag: item.inject_flag,
    tier: item.tier,
    network_keys: item.network_keys,
    env: item.env,
    resources: item.resources,
  }
}

function normalizeTopologyLink(item: RawTopologyLinkData): TopologyLinkData {
  return {
    from_node_key: item.from_node_key,
    to_node_key: item.to_node_key,
  }
}

function normalizeTopologyPolicy(item: RawTopologyTrafficPolicyData): TopologyTrafficPolicyData {
  return {
    source_node_key: item.source_node_key,
    target_node_key: item.target_node_key,
    action: item.action,
    protocol: item.protocol,
    ports: item.ports,
  }
}

function normalizeChallengeTopology(item: RawChallengeTopologyData): ChallengeTopologyData {
  return {
    id: String(item.id),
    challenge_id: String(item.challenge_id),
    template_id: item.template_id == null ? undefined : String(item.template_id),
    entry_node_key: item.entry_node_key,
    networks: item.networks?.map(normalizeTopologyNetwork),
    nodes: item.nodes.map(normalizeTopologyNode),
    links: item.links?.map(normalizeTopologyLink),
    policies: item.policies?.map(normalizeTopologyPolicy),
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function normalizeEnvironmentTemplate(item: RawEnvironmentTemplateData): EnvironmentTemplateData {
  return {
    id: String(item.id),
    name: item.name,
    description: item.description,
    entry_node_key: item.entry_node_key,
    networks: item.networks?.map(normalizeTopologyNetwork),
    nodes: item.nodes.map(normalizeTopologyNode),
    links: item.links?.map(normalizeTopologyLink),
    policies: item.policies?.map(normalizeTopologyPolicy),
    usage_count: item.usage_count,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function normalizeImage(item: RawImageItem): AdminImageListItem {
  return {
    id: String(item.id),
    name: item.name,
    tag: item.tag,
    description: item.description,
    status: item.status,
    size_bytes: item.size,
    created_at: item.created_at,
    updated_at: item.updated_at,
  }
}

function serializeContestPayload(data: AdminContestCreatePayload | AdminContestUpdatePayload) {
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

function normalizeAdminUser(item: RawAdminUser): AdminUserListItem {
  return {
    id: String(item.id),
    username: item.username,
    name: item.name || undefined,
    email: item.email || undefined,
    student_no: item.student_no || undefined,
    teacher_no: item.teacher_no || undefined,
    class_name: item.class_name || undefined,
    status: item.status,
    roles: item.roles,
    created_at: item.created_at,
  }
}

function normalizeCheatDetection(data: RawCheatDetectionData): AdminCheatDetectionData {
  return {
    generated_at: data.generated_at,
    summary: data.summary,
    suspects: data.suspects.map((item) => ({
      ...item,
      user_id: String(item.user_id),
    })),
    shared_ips: data.shared_ips,
  }
}

export async function getDashboard(): Promise<AdminDashboardData> {
  return request<AdminDashboardData>({ method: 'GET', url: '/admin/dashboard' })
}

export async function publishAdminNotification(
  data: AdminNotificationPublishPayload
): Promise<AdminNotificationPublishResult> {
  const response = await request<RawAdminNotificationPublishResult>({
    method: 'POST',
    url: '/admin/notifications',
    data,
  })

  return {
    batch_id: String(response.batch_id),
    recipient_count: response.recipient_count,
  }
}

export async function getUsers(params?: UserListParams): Promise<PageResult<AdminUserListItem>> {
  const response = await request<PageResult<RawAdminUser>>({
    method: 'GET',
    url: '/admin/users',
    params: {
      page: params?.page,
      page_size: params?.page_size,
      keyword: params?.keyword,
      student_no: params?.student_no,
      teacher_no: params?.teacher_no,
      role: params?.role,
      status: params?.status,
      class_name: params?.class_name,
    },
  })

  return {
    ...response,
    list: response.list.map(normalizeAdminUser),
  }
}

export async function createUser(data: AdminUserCreatePayload): Promise<AdminUserUpsertData> {
  const response = await request<{ user: RawAdminUser }>({
    method: 'POST',
    url: '/admin/users',
    data,
  })
  return {
    user: normalizeAdminUser(response.user),
  }
}

export async function updateUser(
  id: string,
  data: AdminUserUpdatePayload
): Promise<AdminUserUpsertData> {
  const response = await request<{ user: RawAdminUser }>({
    method: 'PUT',
    url: `/admin/users/${encodeURIComponent(id)}`,
    data,
  })
  return {
    user: normalizeAdminUser(response.user),
  }
}

export async function deleteUser(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/users/${encodeURIComponent(id)}` })
}

export async function importUsers(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request<AdminUserImportData>({
    method: 'POST',
    url: '/admin/users/import',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function getChallenges(params?: Record<string, unknown>) {
  const response = await request<PageResult<RawAdminChallengeItem>>({
    method: 'GET',
    url: '/authoring/challenges',
    params,
  })
  return {
    ...response,
    list: response.list.map((item) => normalizeChallenge(item)),
  }
}

export async function getChallengeDetail(id: string) {
  const [challenge, flagConfig] = await Promise.all([
    request<RawAdminChallengeItem>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}`,
    }),
    request<RawChallengeFlagConfig>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}/flag`,
    }).catch(() => undefined),
  ])

  return normalizeChallenge(challenge, flagConfig)
}

export async function previewChallengeImport(file: File): Promise<AdminChallengeImportPreview> {
  const form = new FormData()
  form.append('file', file)

  const response = await request<RawChallengeImportPreview>({
    method: 'POST',
    url: '/authoring/challenge-imports',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' },
    suppressErrorToast: true,
  })
  return normalizeChallengeImportPreview(response)
}

export async function listChallengeImports(): Promise<AdminChallengeImportPreview[]> {
  const response = await request<RawChallengeImportPreview[]>({
    method: 'GET',
    url: '/authoring/challenge-imports',
  })
  return Array.isArray(response) ? response.map(normalizeChallengeImportPreview) : []
}

export async function getChallengeImport(id: string): Promise<AdminChallengeImportPreview> {
  const response = await request<RawChallengeImportPreview>({
    method: 'GET',
    url: `/authoring/challenge-imports/${encodeURIComponent(id)}`,
  })
  return normalizeChallengeImportPreview(response)
}

export async function commitChallengeImport(id: string): Promise<AdminChallengeImportCommitData> {
  const response = await request<{ challenge: RawAdminChallengeItem }>({
    method: 'POST',
    url: `/authoring/challenge-imports/${encodeURIComponent(id)}/commit`,
  })

  return {
    challenge: normalizeChallenge(response.challenge),
  }
}

export interface AdminChallengePayload {
  title: string
  description?: string
  category: AdminChallengeListItem['category']
  difficulty: Extract<
    AdminChallengeListItem['difficulty'],
    'beginner' | 'easy' | 'medium' | 'hard' | 'insane'
  >
  points: number
  image_id: number
  attachment_url?: string
  hints?: AdminChallengeHint[]
}

export interface AdminChallengeFlagPayload {
  flag_type: 'static' | 'dynamic' | 'regex' | 'manual_review'
  flag?: string
  flag_regex?: string
  flag_prefix?: string
}

export interface AdminTopologyNodeResourcesPayload {
  cpu_quota?: number
  memory_mb?: number
  pids_limit?: number
}

export interface AdminTopologyNetworkPayload {
  key: string
  name: string
  cidr?: string
  internal?: boolean
}

export interface AdminTopologyNodePayload {
  key: string
  name: string
  image_id?: number
  service_port?: number
  inject_flag?: boolean
  tier?: TopologyNodeData['tier']
  network_keys?: string[]
  env?: Record<string, string>
  resources?: AdminTopologyNodeResourcesPayload
}

export interface AdminTopologyLinkPayload {
  from_node_key: string
  to_node_key: string
}

export interface AdminTopologyPolicyPayload {
  source_node_key: string
  target_node_key: string
  action: TopologyTrafficPolicyData['action']
  protocol?: TopologyTrafficPolicyData['protocol']
  ports?: number[]
}

export interface AdminChallengeTopologyPayload {
  template_id?: number
  entry_node_key?: string
  networks?: AdminTopologyNetworkPayload[]
  nodes?: AdminTopologyNodePayload[]
  links?: AdminTopologyLinkPayload[]
  policies?: AdminTopologyPolicyPayload[]
}

export interface AdminEnvironmentTemplatePayload {
  name: string
  description?: string
  entry_node_key: string
  networks?: AdminTopologyNetworkPayload[]
  nodes: AdminTopologyNodePayload[]
  links?: AdminTopologyLinkPayload[]
  policies?: AdminTopologyPolicyPayload[]
}

export async function createChallenge(data: AdminChallengePayload) {
  const response = await request<RawAdminChallengeItem>({
    method: 'POST',
    url: '/authoring/challenges',
    data,
  })
  return {
    challenge: normalizeChallenge(response),
  }
}

export async function updateChallenge(id: string, data: Partial<AdminChallengePayload>) {
  await request<void>({
    method: 'PUT',
    url: `/authoring/challenges/${encodeURIComponent(id)}`,
    data,
  })
}

export async function configureChallengeFlag(id: string, data: AdminChallengeFlagPayload) {
  return request<{ message: string }>({
    method: 'PUT',
    url: `/authoring/challenges/${encodeURIComponent(id)}/flag`,
    data,
  })
}

export async function getChallengeFlagConfig(id: string) {
  return request<RawChallengeFlagConfig>({
    method: 'GET',
    url: `/authoring/challenges/${encodeURIComponent(id)}/flag`,
  })
}

export async function createChallengePublishRequest(
  id: string
): Promise<AdminChallengePublishRequestData> {
  const response = await request<RawAdminChallengePublishRequestData>({
    method: 'POST',
    url: `/authoring/challenges/${encodeURIComponent(id)}/publish-requests`,
  })

  return normalizeChallengePublishRequest(response)
}

export async function getLatestChallengePublishRequest(
  id: string
): Promise<AdminChallengePublishRequestData | null> {
  try {
    const response = await request<RawAdminChallengePublishRequestData>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}/publish-requests/latest`,
      suppressErrorToast: true,
    })

    return normalizeChallengePublishRequest(response)
  } catch (error) {
    if (
      (error instanceof ApiError && error.status === 404) ||
      ((error as { name?: string; status?: number } | undefined)?.name === 'ApiError' &&
        (error as { status?: number }).status === 404)
    ) {
      return null
    }
    throw error
  }
}

export async function deleteChallenge(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/challenges/${encodeURIComponent(id)}`,
    suppressErrorToast: true,
  })
}

export interface AdminChallengeWriteupPayload {
  title: string
  content: string
  visibility: WriteupVisibility
}

export async function getChallengeWriteup(id: string): Promise<AdminChallengeWriteupData | null> {
  try {
    const response = await request<RawAdminChallengeWriteupData>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}/writeup`,
      suppressErrorToast: true,
    })
    return normalizeAdminChallengeWriteup(response)
  } catch (error) {
    if (
      (error instanceof ApiError && error.status === 404) ||
      ((error as { name?: string; status?: number } | undefined)?.name === 'ApiError' &&
        (error as { status?: number }).status === 404)
    ) {
      return null
    }
    throw error
  }
}

export async function saveChallengeWriteup(id: string, data: AdminChallengeWriteupPayload) {
  const response = await request<RawAdminChallengeWriteupData>({
    method: 'PUT',
    url: `/authoring/challenges/${encodeURIComponent(id)}/writeup`,
    data,
  })
  return normalizeAdminChallengeWriteup(response)
}

export async function deleteChallengeWriteup(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/challenges/${encodeURIComponent(id)}/writeup`,
    suppressErrorToast: true,
  })
}

export async function recommendChallengeWriteup(id: string): Promise<AdminChallengeWriteupData> {
  const response = await request<RawAdminChallengeWriteupData>({
    method: 'POST',
    url: `/authoring/challenges/${encodeURIComponent(id)}/writeup/recommend`,
  })
  return normalizeAdminChallengeWriteup(response)
}

export async function unrecommendChallengeWriteup(id: string): Promise<AdminChallengeWriteupData> {
  const response = await request<RawAdminChallengeWriteupData>({
    method: 'DELETE',
    url: `/authoring/challenges/${encodeURIComponent(id)}/writeup/recommend`,
  })
  return normalizeAdminChallengeWriteup(response)
}

export async function getChallengeTopology(id: string): Promise<ChallengeTopologyData | null> {
  try {
    const response = await request<RawChallengeTopologyData>({
      method: 'GET',
      url: `/authoring/challenges/${encodeURIComponent(id)}/topology`,
      suppressErrorToast: true,
    })
    return normalizeChallengeTopology(response)
  } catch (error) {
    if (
      (error instanceof ApiError && error.status === 404) ||
      ((error as { name?: string; status?: number } | undefined)?.name === 'ApiError' &&
        (error as { status?: number }).status === 404)
    ) {
      return null
    }
    throw error
  }
}

export async function saveChallengeTopology(id: string, data: AdminChallengeTopologyPayload) {
  const response = await request<RawChallengeTopologyData>({
    method: 'PUT',
    url: `/authoring/challenges/${encodeURIComponent(id)}/topology`,
    data,
  })
  return normalizeChallengeTopology(response)
}

export async function deleteChallengeTopology(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/challenges/${encodeURIComponent(id)}/topology`,
    suppressErrorToast: true,
  })
}

export async function getEnvironmentTemplates(
  keyword?: string
): Promise<EnvironmentTemplateData[]> {
  const response = await request<RawEnvironmentTemplateData[]>({
    method: 'GET',
    url: '/authoring/environment-templates',
    params: { keyword },
  })
  return response.map(normalizeEnvironmentTemplate)
}

export async function getEnvironmentTemplate(id: string): Promise<EnvironmentTemplateData> {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'GET',
    url: `/authoring/environment-templates/${encodeURIComponent(id)}`,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function createEnvironmentTemplate(data: AdminEnvironmentTemplatePayload) {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'POST',
    url: '/authoring/environment-templates',
    data,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function updateEnvironmentTemplate(id: string, data: AdminEnvironmentTemplatePayload) {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'PUT',
    url: `/authoring/environment-templates/${encodeURIComponent(id)}`,
    data,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function deleteEnvironmentTemplate(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/environment-templates/${encodeURIComponent(id)}`,
    suppressErrorToast: true,
  })
}

export async function getImages(params?: Record<string, unknown>) {
  const response = await request<PageResult<RawImageItem>>({
    method: 'GET',
    url: '/authoring/images',
    params,
  })
  return {
    ...response,
    list: response.list.map(normalizeImage),
  }
}

export interface AdminImagePayload {
  name: string
  tag: string
  description?: string
}

export async function createImage(data: AdminImagePayload) {
  const response = await request<RawImageItem>({ method: 'POST', url: '/authoring/images', data })
  return {
    image: normalizeImage(response),
  }
}

export async function deleteImage(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/authoring/images/${encodeURIComponent(id)}`,
    suppressErrorToast: true,
  })
}

export async function getAuditLogs(params?: Record<string, unknown>) {
  return request<PageResult<AuditLogItem>>({ method: 'GET', url: '/admin/audit-logs', params })
}

export async function getCheatDetection(): Promise<AdminCheatDetectionData> {
  const response = await request<RawCheatDetectionData>({
    method: 'GET',
    url: '/admin/cheat-detection',
  })
  return normalizeCheatDetection(response)
}

export async function getContests(
  params?: ContestListParams
): Promise<PageResult<ContestDetailData>> {
  const response = await request<PageResult<RawContestItem>>({
    method: 'GET',
    url: '/admin/contests',
    params: {
      page: params?.page,
      page_size: params?.page_size,
      status: serializeContestStatus(params?.status),
    },
  })

  return {
    ...response,
    list: response.list.map(normalizeContest),
  }
}

export async function createContest(
  data: AdminContestCreatePayload
): Promise<{ contest: ContestDetailData }> {
  const contest = await request<RawContestItem>({
    method: 'POST',
    url: '/admin/contests',
    data: serializeContestPayload(data),
  })

  return { contest: normalizeContest(contest) }
}

export async function updateContest(
  id: string,
  data: AdminContestUpdatePayload,
  options?: AdminRequestOptions
): Promise<{ contest: ContestDetailData }> {
  const contest = await request<RawContestItem>({
    method: 'PUT',
    url: `/admin/contests/${encodeURIComponent(id)}`,
    data: serializeContestPayload(data),
    suppressErrorToast: options?.suppressErrorToast,
  })

  return { contest: normalizeContest(contest) }
}

export async function listContestAWDRounds(contestId: string): Promise<AWDRoundData[]> {
  const response = await request<RawAWDRoundItem[]>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds`,
  })
  return response.map(normalizeAWDRound)
}

export async function createContestAWDRound(
  contestId: string,
  data: AdminAWDRoundCreatePayload
): Promise<AWDRoundData> {
  const response = await request<RawAWDRoundItem>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds`,
    data,
    suppressErrorToast: true,
  })
  return normalizeAWDRound(response)
}

export async function getContestAWDReadiness(contestId: string): Promise<AWDReadinessData> {
  const response = await request<RawAWDReadinessData>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/readiness`,
    suppressErrorToast: true,
  })
  return normalizeAWDReadiness(response)
}

export async function listContestTeams(contestId: string): Promise<AdminContestTeamData[]> {
  const response = await request<RawAdminContestTeamItem[]>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(contestId)}/teams`,
  })
  return response.map(normalizeAdminContestTeam)
}

export async function listAdminContestChallenges(
  contestId: string
): Promise<AdminContestChallengeData[]> {
  const response = await request<RawAdminContestChallengeItem[]>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/challenges`,
  })
  return response.map(normalizeAdminContestChallenge)
}

export async function createAdminContestChallenge(
  contestId: string,
  data: AdminContestChallengeCreatePayload
): Promise<AdminContestChallengeData> {
  const response = await request<RawAdminContestChallengeItem>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/challenges`,
    data,
  })
  return normalizeAdminContestChallenge(response)
}

export async function updateAdminContestChallenge(
  contestId: string,
  challengeId: string,
  data: AdminContestChallengeUpdatePayload
): Promise<void> {
  await request<void>({
    method: 'PUT',
    url: `/admin/contests/${encodeURIComponent(contestId)}/challenges/${encodeURIComponent(challengeId)}`,
    data,
  })
}

export async function listContestAWDRoundServices(
  contestId: string,
  roundId: string
): Promise<AWDTeamServiceData[]> {
  const response = await request<RawAWDTeamServiceItem[]>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/services`,
  })
  return response.map(normalizeAWDTeamService)
}

export async function listContestAWDRoundAttacks(
  contestId: string,
  roundId: string
): Promise<AWDAttackLogData[]> {
  const response = await request<RawAWDAttackLogItem[]>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/attacks`,
  })
  return response.map(normalizeAWDAttackLog)
}

export async function getContestAWDRoundSummary(
  contestId: string,
  roundId: string
): Promise<AWDRoundSummaryData> {
  const response = await request<RawAWDRoundSummaryData>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/summary`,
  })
  return normalizeAWDRoundSummary(response)
}

export async function getContestAWDRoundTrafficSummary(
  contestId: string,
  roundId: string
): Promise<AWDTrafficSummaryData> {
  const response = await request<RawAWDTrafficSummaryData>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/traffic/summary`,
  })
  return normalizeAWDTrafficSummary(response)
}

export async function listContestAWDRoundTrafficEvents(
  contestId: string,
  roundId: string,
  params?: AdminAWDTrafficEventsParams
): Promise<AWDTrafficEventPageData> {
  const response = await request<PageResult<RawAWDTrafficEventItem>>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/traffic/events`,
    params,
  })
  return {
    ...response,
    list: response.list.map(normalizeAWDTrafficEvent),
  }
}

export async function getAdminContestLiveScoreboard(
  contestId: string,
  params?: Record<string, unknown>
): Promise<ContestScoreboardData> {
  const response = await request<ContestScoreboardData>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/scoreboard/live`,
    params,
  })
  return normalizeContestScoreboard(response)
}

export async function runContestAWDCurrentRoundCheck(
  contestId: string,
  data?: AdminAWDCurrentRoundCheckPayload
): Promise<AWDCheckerRunData> {
  const response = await request<RawAWDCheckerRunData>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/current-round/check`,
    data,
    suppressErrorToast: true,
  })
  return normalizeAWDCheckerRun(response)
}

export async function runContestAWDRoundCheck(
  contestId: string,
  roundId: string
): Promise<AWDCheckerRunData> {
  const response = await request<RawAWDCheckerRunData>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/check`,
  })
  return normalizeAWDCheckerRun(response)
}

export async function runContestAWDCheckerPreview(
  contestId: string,
  data: AdminAWDCheckerPreviewPayload
): Promise<AWDCheckerPreviewData> {
  const response = await request<RawAWDCheckerPreviewData>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/checker-preview`,
    data,
  })
  return normalizeAWDCheckerPreview(response)
}

export async function createContestAWDServiceCheck(
  contestId: string,
  roundId: string,
  data: AdminAWDServiceCheckPayload
): Promise<AWDTeamServiceData> {
  const response = await request<RawAWDTeamServiceItem>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/services/check`,
    data,
  })
  return normalizeAWDTeamService(response)
}

export async function createContestAWDAttackLog(
  contestId: string,
  roundId: string,
  data: AdminAWDAttackLogPayload
): Promise<AWDAttackLogData> {
  const response = await request<RawAWDAttackLogItem>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/rounds/${encodeURIComponent(roundId)}/attacks`,
    data,
  })
  return normalizeAWDAttackLog(response)
}
