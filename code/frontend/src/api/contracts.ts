import type { UserRole } from '@/utils/constants'

export type ID = string
export type ISODateTime = string

export interface FieldError {
  field: string
  message: string
}

export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
  request_id: string
  errors?: FieldError[]
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface AuthUser {
  id: ID
  username: string
  role: UserRole
  avatar?: string
  name?: string
  class_name?: string
}

export type ChallengeCategory = 'web' | 'pwn' | 'reverse' | 'crypto' | 'misc' | 'forensics'
export type ChallengeDifficulty = 'beginner' | 'easy' | 'medium' | 'hard' | 'insane'
export type InstanceSharing = 'per_user' | 'per_team' | 'shared'
export type InstanceAccessProtocol = 'http' | 'tcp'

export interface InstanceAccessInfo {
  protocol: InstanceAccessProtocol
  host?: string
  port?: number
  command?: string
}

export interface ChallengeListItem {
  id: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  tags: string[]
  solved_count: number
  total_attempts: number
  is_solved: boolean
  points: number
  created_at: ISODateTime
}

export interface ChallengeHint {
  id: ID
  level: number
  title?: string
  content?: string
}

export interface ChallengeDetailData {
  id: ID
  title: string
  description: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  tags: string[]
  points: number
  need_target: boolean
  flag_type?: FlagType
  instance_sharing: InstanceSharing
  attachment_url?: string
  is_solved: boolean
  solved_at?: ISODateTime
  hints: ChallengeHint[]
}

export interface ChallengeWriteupData {
  id: ID
  challenge_id: ID
  title: string
  content: string
  visibility: 'private' | 'public'
  requires_spoiler_warning: boolean
  is_recommended?: boolean
  recommended_at?: ISODateTime
  recommended_by?: ID
  created_at: ISODateTime
  updated_at: ISODateTime
}

export type SubmissionWriteupStatus = 'draft' | 'published' | 'submitted'
export type SubmissionWriteupVisibilityStatus = 'visible' | 'hidden'

export interface SubmissionWriteupData {
  id: ID
  user_id: ID
  challenge_id: ID
  contest_id?: ID
  title: string
  content: string
  submission_status: SubmissionWriteupStatus
  visibility_status?: SubmissionWriteupVisibilityStatus
  is_recommended?: boolean
  recommended_at?: ISODateTime
  recommended_by?: ID
  published_at?: ISODateTime
  created_at: ISODateTime
  updated_at: ISODateTime
}

export interface RecommendedChallengeSolutionData {
  id: ID
  source_type: 'official' | 'community'
  source_id: ID
  challenge_id: ID
  title: string
  content: string
  author_name: string
  is_recommended: boolean
  recommended_at?: ISODateTime
  updated_at: ISODateTime
}

export interface CommunityChallengeSolutionData {
  id: ID
  challenge_id: ID
  user_id: ID
  title: string
  content: string
  content_preview: string
  author_name: string
  submission_status: SubmissionWriteupStatus
  visibility_status: SubmissionWriteupVisibilityStatus
  is_recommended: boolean
  published_at?: ISODateTime
  updated_at: ISODateTime
}

export interface TeacherSubmissionWriteupItemData {
  id: ID
  user_id: ID
  student_username: string
  student_name?: string
  student_no?: string
  class_name?: string
  challenge_id: ID
  challenge_title: string
  title: string
  content_preview: string
  submission_status: SubmissionWriteupStatus
  visibility_status: SubmissionWriteupVisibilityStatus
  is_recommended: boolean
  published_at?: ISODateTime
  updated_at: ISODateTime
}

export interface TeacherSubmissionWriteupDetailData extends SubmissionWriteupData {
  student_username: string
  student_name?: string
  student_no?: string
  class_name?: string
  challenge_title: string
}

export type InstanceStatus =
  | 'pending'
  | 'creating'
  | 'running'
  | 'expired'
  | 'destroying'
  | 'destroyed'
  | 'failed'
  | 'crashed'
export type FlagType = 'static' | 'dynamic' | 'regex' | 'manual_review'

export interface InstanceData {
  id: ID
  contest_mode?: ContestMode
  challenge_id: ID
  awd_challenge_id?: ID
  status: InstanceStatus
  share_scope: InstanceSharing
  access_url?: string
  access?: InstanceAccessInfo
  ssh_info?: { host: string; port: number; username: string }
  flag_type: FlagType
  expires_at: ISODateTime
  remaining_extends: number
  created_at: ISODateTime
  queue_position?: number
  eta_seconds?: number
  progress?: number
}

export interface InstanceExtendData {
  id: ID
  expires_at: ISODateTime
  remaining_extends: number
}

export interface InstanceListItem extends InstanceData {
  challenge_title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
}

export interface SubmitFlagData {
  is_correct: boolean
  status: 'correct' | 'incorrect' | 'pending_review'
  message?: string
  points?: number
  submitted_at: ISODateTime
  instance_shutdown_at?: ISODateTime
}

export interface ChallengeSubmissionRecordData {
  id: ID
  status: 'correct' | 'incorrect' | 'pending_review'
  answer?: string
  submitted_at: ISODateTime
}

export type TeacherManualReviewStatus = 'pending' | 'approved' | 'rejected'

export interface TeacherManualReviewSubmissionItemData {
  id: ID
  user_id: ID
  student_username: string
  student_name?: string
  class_name?: string
  challenge_id: ID
  challenge_title: string
  answer_preview: string
  review_status: TeacherManualReviewStatus
  submitted_at: ISODateTime
  reviewed_at?: ISODateTime
  updated_at: ISODateTime
}

export interface TeacherManualReviewSubmissionDetailData {
  id: ID
  user_id: ID
  student_username: string
  student_name?: string
  class_name?: string
  challenge_id: ID
  challenge_title: string
  answer: string
  is_correct: boolean
  score: number
  review_status: TeacherManualReviewStatus
  reviewed_at?: ISODateTime
  review_comment?: string
  submitted_at: ISODateTime
  updated_at: ISODateTime
  reviewer_name?: string
}

export interface MyProgressData {
  total_score?: number
  total_solved?: number
  rank?: number
  category_stats?: Array<{ category: string; total: number; solved: number }>
  difficulty_stats?: Array<{ difficulty: string; total: number; solved: number }>
  total_challenges?: number
  solved_challenges?: number
  by_category?: Record<string, { total: number; solved: number }>
  by_difficulty?: Record<string, { total: number; solved: number }>
}

export interface TimelineEvent {
  id: ID
  type: 'solve' | 'submit' | 'instance' | 'hint' | string
  title: string
  detail?: string
  created_at: ISODateTime
  challenge_id?: ID
  is_correct?: boolean
  points?: number
  meta?: Record<string, unknown>
}

export type ContestMode = 'jeopardy' | 'awd' | 'awd_plus' | 'king_of_hill'
export type ContestStatus =
  | 'draft'
  | 'published'
  | 'registering'
  | 'running'
  | 'frozen'
  | 'ended'
  | 'cancelled'
  | 'archived'

export interface ContestListItem {
  id: ID
  title: string
  mode: ContestMode
  status: ContestStatus
  starts_at: ISODateTime
  ends_at: ISODateTime
  register_ends_at?: ISODateTime
  scoreboard_frozen?: boolean
}

export interface ContestDetailData extends ContestListItem {
  description?: string
  rules?: string
  team_size_limit?: number
}

export interface ContestChallengeItem {
  id: ID
  challenge_id: ID
  awd_challenge_id?: ID
  awd_service_id?: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  points: number
  solved_count: number
  is_solved: boolean
}

export interface ScoreboardRow {
  rank: number
  team_id: ID
  team_name: string
  score: number
  solved_count: number
  last_submission_at?: ISODateTime
}

export interface PracticeRankingItemData {
  rank: number
  user_id: ID
  username: string
  total_score: number
  solved_count: number
  class_name?: string
}

export interface ContestScoreboardData {
  contest: {
    id: ID
    title: string
    status: ContestStatus
    started_at: ISODateTime
    ends_at: ISODateTime
  }
  scoreboard: PageResult<ScoreboardRow>
  frozen: boolean
}

export interface ContestAnnouncement {
  id: ID
  title: string
  content?: string
  created_at: ISODateTime
}

export interface TeamData {
  id: ID
  name: string
  invite_code?: string
  captain_user_id: ID
  members: Array<{ user_id: ID; username: string; joined_at: ISODateTime }>
}

export interface ContestMyProgressData {
  contest_id: ID
  team_id?: ID
  solved: Array<{ contest_challenge_id: ID; solved_at: ISODateTime; points_earned: number }>
}

export type AWDRoundStatus = 'pending' | 'running' | 'finished'
export type AWDServiceStatus = 'up' | 'down' | 'compromised'
export type AWDAttackType = 'flag_capture' | 'service_exploit'
export type AWDAttackSource = 'legacy' | 'manual_attack_log' | 'submission'
export type AWDCheckerType = 'legacy_probe' | 'http_standard' | 'tcp_standard' | 'script_checker'
export type AWDReadinessAction = 'create_round' | 'run_current_round_check' | 'start_contest'
export type AWDReadinessBlockingReason =
  | 'missing_checker'
  | 'invalid_checker_config'
  | 'pending_validation'
  | 'last_preview_failed'
  | 'validation_stale'
export type AWDReadinessGlobalReason = 'no_challenges'

export interface AWDRoundData {
  id: ID
  contest_id: ID
  round_number: number
  status: AWDRoundStatus
  started_at?: ISODateTime
  ended_at?: ISODateTime
  attack_score: number
  defense_score: number
  created_at: ISODateTime
  updated_at: ISODateTime
}

export interface AWDTeamServiceData {
  id: ID
  round_id: ID
  team_id: ID
  team_name: string
  service_id?: ID
  service_name?: string
  awd_challenge_id: ID
  awd_challenge_title?: string
  service_status: AWDServiceStatus
  checker_type?: AWDCheckerType
  check_result: Record<string, unknown>
  attack_received: number
  sla_score: number
  defense_score: number
  attack_score: number
  updated_at: ISODateTime
}

export interface AWDAttackLogData {
  id: ID
  round_id: ID
  attacker_team_id: ID
  attacker_team: string
  victim_team_id: ID
  victim_team: string
  service_id?: ID
  awd_challenge_id: ID
  attack_type: AWDAttackType
  source: AWDAttackSource
  submitted_flag?: string
  is_success: boolean
  score_gained: number
  created_at: ISODateTime
}

export type ContestAWDWorkspaceEventDirection = 'attack_in' | 'attack_out'

export interface ContestAWDWorkspaceTeamData {
  team_id: ID
  team_name: string
}

export interface ContestAWDWorkspaceServiceData {
  service_id?: ID
  awd_challenge_id: ID
  instance_id?: ID
  instance_status?: InstanceStatus
  access_url?: string
  service_status?: AWDServiceStatus
  operation_status?: 'requested' | 'provisioning' | 'recovering' | 'recovered' | 'succeeded' | 'failed'
  operation_type?: 'start' | 'restart' | 'recover' | 'recreate'
  operation_reason?: string
  operation_sla_billable?: boolean
  checker_type?: AWDCheckerType
  attack_received: number
  sla_score: number
  defense_score: number
  attack_score: number
  updated_at?: ISODateTime
}

export interface ContestAWDWorkspaceTargetServiceData {
  service_id?: ID
  awd_challenge_id: ID
  reachable: boolean
}

export interface ContestAWDWorkspaceTargetTeamData {
  team_id: ID
  team_name: string
  services: ContestAWDWorkspaceTargetServiceData[]
}

export interface ContestAWDWorkspaceRecentEventData {
  id: ID
  direction: ContestAWDWorkspaceEventDirection
  service_id?: ID
  awd_challenge_id: ID
  peer_team_id: ID
  peer_team_name: string
  is_success: boolean
  score_gained: number
  created_at: ISODateTime
}

export interface ContestAWDWorkspaceData {
  contest_id: ID
  current_round?: AWDRoundData
  my_team?: ContestAWDWorkspaceTeamData | null
  services: ContestAWDWorkspaceServiceData[]
  targets: ContestAWDWorkspaceTargetTeamData[]
  recent_events: ContestAWDWorkspaceRecentEventData[]
}

export interface SSHProfileData {
  alias: string
  host_name: string
  port: number
  user: string
}

export interface AWDDefenseSSHAccessData {
  host: string
  port: number
  username: string
  password: string
  command: string
  ssh_profile?: SSHProfileData
  expires_at: ISODateTime
}

export interface AWDRoundSummaryItemData {
  team_id: ID
  team_name: string
  service_up_count: number
  service_down_count: number
  service_compromised_count: number
  sla_score: number
  defense_score: number
  attack_score: number
  successful_attack_count: number
  successful_breach_count: number
  unique_attackers_against: number
  total_score: number
}

export interface AWDRoundMetricsData {
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

export interface AWDRoundSummaryData {
  round: AWDRoundData
  metrics?: AWDRoundMetricsData
  items: AWDRoundSummaryItemData[]
}

export type AWDTrafficStatusGroup = 'success' | 'redirect' | 'client_error' | 'server_error'

export interface AWDTrafficTopTeamData {
  team_id: ID
  team_name: string
  request_count: number
  error_count: number
}

export interface AWDTrafficTopChallengeData {
  awd_challenge_id: ID
  awd_challenge_title?: string
  request_count: number
  error_count: number
}

export interface AWDTrafficTopPathData {
  path: string
  request_count: number
  error_count: number
  last_status_code: number
}

export interface AWDTrafficTrendBucketData {
  bucket_start_at: ISODateTime
  bucket_end_at?: ISODateTime
  request_count: number
  error_count: number
}

export interface AWDTrafficSummaryData {
  round?: AWDRoundData
  contest_id: ID
  round_id: ID
  total_request_count: number
  active_attacker_team_count: number
  victim_team_count: number
  unique_path_count: number
  error_request_count: number
  latest_event_at?: ISODateTime
  top_attackers: AWDTrafficTopTeamData[]
  top_victims: AWDTrafficTopTeamData[]
  top_challenges: AWDTrafficTopChallengeData[]
  top_paths: AWDTrafficTopPathData[]
  top_error_paths: AWDTrafficTopPathData[]
  trend_buckets: AWDTrafficTrendBucketData[]
}

export interface AWDTrafficEventData {
  id: ID
  contest_id: ID
  round_id: ID
  attacker_team_id: ID
  attacker_team_name?: string
  victim_team_id: ID
  victim_team_name?: string
  service_id?: ID
  awd_challenge_id: ID
  awd_challenge_title?: string
  method: string
  path: string
  status_code: number
  status_group: AWDTrafficStatusGroup
  is_error: boolean
  source: string
  request_id?: string
  occurred_at: ISODateTime
}

export type AWDTrafficEventPageData = PageResult<AWDTrafficEventData>

export interface AWDCheckerRunData {
  round: AWDRoundData
  services: AWDTeamServiceData[]
}

export interface AWDCheckerPreviewContextData {
  access_url: string
  preview_flag: string
  round_number: number
  team_id: ID
  awd_challenge_id: ID
}

export type AWDCheckerValidationState = 'pending' | 'passed' | 'failed' | 'stale'

export interface AWDCheckerPreviewData {
  checker_type?: AWDCheckerType
  service_status: AWDTeamServiceData['service_status']
  check_result: Record<string, unknown>
  preview_context: AWDCheckerPreviewContextData
  preview_token?: string
}

export interface AWDReadinessItemData {
  awd_challenge_id: ID
  title: string
  checker_type?: AWDCheckerType
  validation_state: 'pending' | 'failed' | 'stale' | 'passed'
  last_preview_at?: ISODateTime
  last_access_url?: string
  blocking_reason: AWDReadinessBlockingReason
}

export interface AWDReadinessData {
  contest_id: ID
  ready: boolean
  total_challenges: number
  passed_challenges: number
  pending_challenges: number
  failed_challenges: number
  stale_challenges: number
  missing_checker_challenges: number
  blocking_count: number
  global_blocking_reasons: AWDReadinessGlobalReason[]
  blocking_actions: AWDReadinessAction[]
  items: AWDReadinessItemData[]
}

export interface AdminContestTeamData {
  id: ID
  contest_id: ID
  name: string
  captain_id: ID
  invite_code?: string
  max_members: number
  member_count: number
  created_at: ISODateTime
}

export interface AdminContestChallengeRelationData {
  id: ID
  contest_id: ID
  challenge_id: ID
  title?: string
  category?: ChallengeCategory
  difficulty?: ChallengeDifficulty
  points: number
  order: number
  is_visible: boolean
  created_at: ISODateTime
}

export interface AdminContestChallengeViewData extends AdminContestChallengeRelationData {
  awd_service_id?: ID
  awd_challenge_id?: ID
  awd_service_display_name?: string
  awd_checker_type?: AWDCheckerType
  awd_checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_validation_state?: AWDCheckerValidationState
  awd_checker_last_preview_at?: ISODateTime
  awd_checker_last_preview_result?: AWDCheckerPreviewData
}

export type AdminContestChallengeData = AdminContestChallengeViewData

export interface AdminContestAWDServiceData {
  id: ID
  contest_id: ID
  awd_challenge_id: ID
  title?: string
  category?: ChallengeCategory
  difficulty?: ChallengeDifficulty
  display_name: string
  order: number
  is_visible: boolean
  score_config?: Record<string, unknown>
  runtime_config?: Record<string, unknown>
  checker_type?: AWDCheckerType
  checker_config?: Record<string, unknown>
  sla_score?: number
  defense_score?: number
  validation_state?: AWDCheckerValidationState
  last_preview_at?: ISODateTime
  last_preview_result?: AWDCheckerPreviewData
  created_at: ISODateTime
  updated_at: ISODateTime
}

export interface AdminContestAWDInstanceTeamData {
  team_id: ID
  team_name: string
  captain_id: ID
}

export interface AdminContestAWDInstanceServiceData {
  service_id: ID
  awd_challenge_id: ID
  display_name: string
  is_visible: boolean
}

export interface AdminContestAWDInstanceItemData {
  team_id: ID
  service_id: ID
  instance?: InstanceData
}

export interface AdminContestAWDInstanceOrchestrationData {
  contest_id: ID
  teams: AdminContestAWDInstanceTeamData[]
  services: AdminContestAWDInstanceServiceData[]
  instances: AdminContestAWDInstanceItemData[]
}

export type NotificationType = 'system' | 'contest' | 'challenge' | 'team'

export interface NotificationItem {
  id: ID
  type: NotificationType
  title: string
  content?: string
  link?: string
  level?: 'info' | 'success' | 'warning' | 'error'
  unread: boolean
  created_at: ISODateTime
}

export type NotificationAudienceRule =
  | { type: 'all' }
  | { type: 'role'; values: UserRole[] }
  | { type: 'class'; values: string[] }
  | { type: 'user'; values: string[] }

export interface NotificationAudienceRules {
  mode: 'union'
  rules: NotificationAudienceRule[]
}

export interface AdminNotificationPublishPayload {
  type: NotificationType
  title: string
  content: string
  link?: string
  audience_rules: NotificationAudienceRules
}

export interface AdminNotificationPublishResult {
  batch_id: ID
  recipient_count: number
}

export interface TeacherClassItem {
  name: string
  student_count?: number
}

export interface TeacherClassSummaryData {
  class_name: string
  student_count: number
  average_solved: number
  active_student_count: number
  active_rate: number
  recent_event_count: number
}

export interface TeacherClassTrendPoint {
  date: string
  active_student_count: number
  event_count: number
  solve_count: number
}

export interface TeacherClassTrendData {
  class_name: string
  points: TeacherClassTrendPoint[]
}

export interface TeacherReviewStudentRef {
  id: ID
  username: string
  name?: string
}

export interface TeacherClassReviewItemData {
  key: string
  title: string
  detail: string
  accent: 'danger' | 'warning' | 'success' | 'primary'
  students?: TeacherReviewStudentRef[]
  recommendation?: RecommendationItem
}

export interface TeacherClassReviewData {
  class_name: string
  items: TeacherClassReviewItemData[]
}

export interface TeacherStudentItem {
  id: ID
  username: string
  student_no?: string
  name?: string
  class_name?: string
  solved_count?: number
  total_score?: number
  recent_event_count?: number
  weak_dimension?: string
  progress?: MyProgressData
}

export interface TeacherEvidenceSummaryData {
  total_events: number
  proxy_request_count: number
  submit_count: number
  success_count: number
  challenge_id: ID
}

export interface TeacherEvidenceEventData {
  type: string
  challenge_id: ID
  title: string
  detail: string
  timestamp: ISODateTime
  meta?: Record<string, unknown>
}

export interface TeacherEvidenceData {
  summary: TeacherEvidenceSummaryData
  events: TeacherEvidenceEventData[]
}

export interface TeacherAttackActorData {
  user_id: ID
  team_id?: ID
}

export interface TeacherAttackTargetData {
  challenge_id?: ID
  contest_id?: ID
  round_id?: ID
  service_id?: ID
  victim_team_id?: ID
}

export interface TeacherAttackEventData {
  id: ID
  session_id?: ID
  type: string
  stage: string
  source: string
  occurred_at: ISODateTime
  actor: TeacherAttackActorData
  target: TeacherAttackTargetData
  summary: string
  meta?: Record<string, unknown>
  capture_available: boolean
  capture_ref?: Record<string, unknown>
}

export interface TeacherAttackSessionData {
  id: ID
  mode: 'practice' | 'jeopardy' | 'awd' | string
  student_id: ID
  team_id?: ID
  challenge_id?: ID
  contest_id?: ID
  round_id?: ID
  service_id?: ID
  victim_team_id?: ID
  title: string
  started_at: ISODateTime
  ended_at: ISODateTime
  result: 'success' | 'failed' | 'in_progress' | 'unknown' | string
  event_count: number
  capture_count: number
  events?: TeacherAttackEventData[]
}

export interface TeacherAttackSessionSummaryData {
  total_sessions: number
  success_count: number
  failed_count: number
  in_progress_count: number
  unknown_count: number
  event_count: number
  capture_available_count: number
}

export interface TeacherAttackSessionResponseData {
  summary: TeacherAttackSessionSummaryData
  sessions: TeacherAttackSessionData[]
}

export interface TeacherInstanceItem {
  id: ID
  student_id: ID
  student_name: string
  student_username: string
  student_no?: string
  class_name: string
  challenge_id: ID
  challenge_title: string
  status: string
  access_url?: string
  expires_at: ISODateTime
  remaining_time: number
  extend_count: number
  max_extends: number
  created_at: ISODateTime
}

export interface SkillDimensionScore {
  key: string
  name: string
  value: number
}

export interface SkillProfileData {
  dimensions: SkillDimensionScore[]
  updated_at?: ISODateTime
}

export interface RecommendationItem {
  challenge_id: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  reason: string
}

export interface ReportExportData {
  report_id: ID
  status: 'ready' | 'processing' | 'failed'
  download_url?: string
  expires_at?: ISODateTime
  error_message?: string
}

export interface TeacherAWDReviewContestItemData {
  id: ID
  title: string
  mode: ContestMode
  status: ContestStatus
  current_round?: number
  round_count: number
  team_count: number
  latest_evidence_at?: ISODateTime
  export_ready: boolean
}

export interface TeacherAWDReviewScopeData {
  snapshot_type: 'live' | 'final' | string
  requested_by: number
  requested_id: ID
  requested_role?: string
}

export interface TeacherAWDReviewOverviewData {
  round_count: number
  team_count: number
  service_count: number
  attack_count: number
  traffic_count: number
  latest_evidence_at?: ISODateTime
}

export interface TeacherAWDReviewRoundItemData {
  id: ID
  contest_id: ID
  round_number: number
  status: string
  started_at?: ISODateTime
  ended_at?: ISODateTime
  attack_score: number
  defense_score: number
  service_count: number
  attack_count: number
  traffic_count: number
}

export interface TeacherAWDReviewTeamItemData {
  team_id: ID
  team_name: string
  captain_id: ID
  total_score: number
  member_count: number
  last_solve_at?: ISODateTime
}

export interface TeacherAWDReviewServiceItemData {
  id: ID
  round_id: ID
  team_id: ID
  team_name: string
  service_id?: ID
  challenge_id: ID
  challenge_title: string
  service_status: string
  attack_received: number
  sla_score: number
  defense_score: number
  attack_score: number
  updated_at: ISODateTime
}

export interface TeacherAWDReviewAttackItemData {
  id: ID
  round_id: ID
  attacker_team_id: ID
  attacker_team_name: string
  victim_team_id: ID
  victim_team_name: string
  service_id?: ID
  challenge_id: ID
  challenge_title: string
  attack_type: string
  source: string
  submitted_flag?: string
  is_success: boolean
  score_gained: number
  created_at: ISODateTime
}

export interface TeacherAWDReviewTrafficItemData {
  id: ID
  contest_id: ID
  round_id: ID
  attacker_team_id: ID
  attacker_team_name: string
  victim_team_id: ID
  victim_team_name: string
  service_id?: ID
  challenge_id: ID
  challenge_title: string
  method: string
  path: string
  status_code: number
  source: string
  created_at: ISODateTime
}

export interface TeacherAWDReviewSelectedRoundData {
  round: TeacherAWDReviewRoundItemData
  teams: TeacherAWDReviewTeamItemData[]
  services: TeacherAWDReviewServiceItemData[]
  attacks: TeacherAWDReviewAttackItemData[]
  traffic: TeacherAWDReviewTrafficItemData[]
}

export interface TeacherAWDReviewArchiveData {
  generated_at: ISODateTime
  scope: TeacherAWDReviewScopeData
  contest: TeacherAWDReviewContestItemData
  overview?: TeacherAWDReviewOverviewData
  rounds: TeacherAWDReviewRoundItemData[]
  selected_round?: TeacherAWDReviewSelectedRoundData
}

export interface ReviewArchiveStudentData {
  id: ID
  username: string
  name?: string
  class_name?: string
}

export interface ReviewArchiveSummaryData {
  total_challenges: number
  total_solved: number
  total_score: number
  rank: number
  total_attempts: number
  timeline_event_count: number
  evidence_event_count: number
  writeup_count: number
  manual_review_count: number
  correct_submission_count: number
  last_activity_at?: ISODateTime
}

export interface ReviewArchiveEvidenceItemData {
  type: string
  challenge_id: ID
  title: string
  detail?: string
  timestamp: ISODateTime
  meta?: Record<string, unknown>
}

export interface ReviewArchiveWriteupItemData {
  id: ID
  challenge_id: ID
  challenge_title: string
  title: string
  submission_status: string
  visibility_status: string
  is_recommended: boolean
  published_at?: ISODateTime
  updated_at: ISODateTime
}

export interface ReviewArchiveManualReviewItemData {
  id: ID
  challenge_id: ID
  challenge_title: string
  answer: string
  review_status: string
  submitted_at: ISODateTime
  reviewed_at?: ISODateTime
  review_comment?: string
  score: number
  reviewer_name?: string
}

export interface ReviewArchiveObservationItemData {
  key: string
  label: string
  level: string
  summary: string
  evidence?: string
}

export interface ReviewArchiveData {
  generated_at: ISODateTime
  student: ReviewArchiveStudentData
  summary: ReviewArchiveSummaryData
  skill_profile: SkillProfileData
  timeline: TimelineEvent[]
  evidence: ReviewArchiveEvidenceItemData[]
  writeups: ReviewArchiveWriteupItemData[]
  manual_reviews: ReviewArchiveManualReviewItemData[]
  teacher_observations: {
    items: ReviewArchiveObservationItemData[]
  }
}

export interface AdminContainerStat {
  container_id: string
  container_name: string
  cpu_percent: number
  memory_percent: number
  memory_usage: number
  memory_limit: number
}

export interface AdminResourceAlert {
  container_id: string
  type: 'cpu' | 'memory'
  value: number
  threshold: number
  message: string
}

export interface AdminDashboardData {
  online_users: number
  active_containers: number
  cpu_usage: number
  memory_usage: number
  container_stats: AdminContainerStat[]
  alerts: AdminResourceAlert[]
}

export type UserStatus = 'active' | 'inactive' | 'locked' | 'banned'

export interface AdminUserListItem {
  id: ID
  username: string
  email?: string
  student_no?: string
  teacher_no?: string
  name?: string
  class_name?: string
  status: UserStatus
  roles: UserRole[]
  created_at: ISODateTime
}

export interface AdminUserUpsertData {
  user: AdminUserListItem
}

export interface AdminUserImportData {
  created: number
  updated: number
  failed: number
  errors?: Array<{ row: number; message: string }>
}

export interface AdminCheatDetectionSummary {
  submit_burst_users: number
  shared_ip_groups: number
  affected_users: number
}

export interface AdminCheatDetectionSuspect {
  user_id: ID
  username: string
  submit_count: number
  last_seen_at: ISODateTime
  reason: string
}

export interface AdminCheatDetectionIPGroup {
  ip: string
  user_count: number
  usernames: string[]
}

export interface AdminCheatDetectionData {
  generated_at: ISODateTime
  summary: AdminCheatDetectionSummary
  suspects: AdminCheatDetectionSuspect[]
  shared_ips: AdminCheatDetectionIPGroup[]
}

export type ChallengeStatus = 'draft' | 'published' | 'archived'
export type WriteupVisibility = 'private' | 'public'

export interface AdminChallengeHint {
  id?: ID
  level: number
  title?: string
  content: string
}

export interface AdminChallengeImportAttachment {
  name: string
  path: string
}

export interface AdminChallengeImportFlag {
  type: Extract<FlagType, 'static' | 'dynamic' | 'regex' | 'manual_review'>
  prefix?: string
}

export interface AdminChallengeImportRuntime {
  type?: string
  image_ref?: string
}

export interface AdminChallengeImportExtensions {
  topology: {
    source?: string
    enabled: boolean
  }
}

export interface ChallengePackageFileData {
  path: string
  size: number
}

export interface AdminChallengeImportTopologyNodeData {
  key: string
  name: string
  image_ref?: string
  service_port?: number
  inject_flag?: boolean
  tier?: TopologyTier
  network_keys?: string[]
  env?: Record<string, string>
}

export interface AdminChallengeImportTopologyData {
  source?: string
  entry_node_key: string
  networks?: TopologyNetworkData[]
  nodes: AdminChallengeImportTopologyNodeData[]
  links?: TopologyLinkData[]
  policies?: TopologyTrafficPolicyData[]
}

export interface AdminChallengeImportPreview {
  id: ID
  file_name: string
  slug: string
  title: string
  description: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  points: number
  attachments?: AdminChallengeImportAttachment[]
  hints?: AdminChallengeHint[]
  flag: AdminChallengeImportFlag
  runtime: AdminChallengeImportRuntime
  extensions: AdminChallengeImportExtensions
  topology?: AdminChallengeImportTopologyData
  package_files?: ChallengePackageFileData[]
  warnings?: string[]
  created_at: ISODateTime
}

export interface AdminChallengeWriteupData {
  id: ID
  challenge_id: ID
  title: string
  content: string
  visibility: WriteupVisibility
  created_by?: ID
  is_recommended: boolean
  recommended_at?: ISODateTime
  recommended_by?: ID
  created_at: ISODateTime
  updated_at: ISODateTime
}

export type TopologyTier = 'public' | 'service' | 'internal'
export type TopologyPolicyAction = 'allow' | 'deny'
export type TopologyPolicyProtocol = 'tcp' | 'udp' | 'any'

export interface TopologyNetworkData {
  key: string
  name: string
  cidr?: string
  internal?: boolean
}

export interface TopologyNodeResourcesData {
  cpu_quota?: number
  memory_mb?: number
  pids_limit?: number
}

export interface TopologyNodeData {
  key: string
  name: string
  image_id?: ID
  service_port?: number
  inject_flag?: boolean
  tier?: TopologyTier
  network_keys?: string[]
  env?: Record<string, string>
  resources?: TopologyNodeResourcesData
}

export interface TopologyLinkData {
  from_node_key: string
  to_node_key: string
}

export interface TopologyTrafficPolicyData {
  source_node_key: string
  target_node_key: string
  action: TopologyPolicyAction
  protocol?: TopologyPolicyProtocol
  ports?: number[]
}

export interface ChallengeTopologyData {
  id: ID
  challenge_id: ID
  template_id?: ID
  entry_node_key: string
  networks?: TopologyNetworkData[]
  nodes: TopologyNodeData[]
  links?: TopologyLinkData[]
  policies?: TopologyTrafficPolicyData[]
  source_type?: 'platform_manual' | 'package_import'
  source_path?: string
  sync_status?: 'clean' | 'drifted'
  package_revision_id?: ID
  last_export_revision_id?: ID
  package_baseline?: TopologySpecData
  package_files?: ChallengePackageFileData[]
  package_revisions?: ChallengePackageRevisionData[]
  created_at: ISODateTime
  updated_at: ISODateTime
}

export interface TopologySpecData {
  entry_node_key: string
  networks?: TopologyNetworkData[]
  nodes: TopologyNodeData[]
  links?: TopologyLinkData[]
  policies?: TopologyTrafficPolicyData[]
}

export interface ChallengePackageRevisionData {
  id: ID
  revision_no: number
  source_type: 'imported' | 'exported'
  parent_revision_id?: ID
  package_slug?: string
  archive_path?: string
  source_dir?: string
  topology_source_path?: string
  created_by?: ID
  created_at: ISODateTime
  updated_at: ISODateTime
}

export interface ChallengePackageExportData {
  challenge_id: ID
  revision_id: ID
  archive_path: string
  source_dir: string
  file_name: string
  download_url?: string
  created_at: ISODateTime
}

export interface EnvironmentTemplateData {
  id: ID
  name: string
  description: string
  entry_node_key: string
  networks?: TopologyNetworkData[]
  nodes: TopologyNodeData[]
  links?: TopologyLinkData[]
  policies?: TopologyTrafficPolicyData[]
  usage_count: number
  created_at: ISODateTime
  updated_at: ISODateTime
}

export type AWDServiceType = 'web_http' | 'binary_tcp' | 'multi_container'
export type AWDDeploymentMode = 'single_container' | 'topology'
export type AWDChallengeStatus = 'draft' | 'published' | 'archived'
export type AWDReadinessStatus = 'pending' | 'passed' | 'failed'

export interface AdminAwdChallengeData {
  id: ID
  name: string
  slug: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  description: string
  service_type: AWDServiceType
  deployment_mode: AWDDeploymentMode
  version: string
  status: AWDChallengeStatus
  readiness_status: AWDReadinessStatus
  checker_type?: AWDCheckerType
  checker_config?: Record<string, unknown>
  flag_mode?: string
  flag_config?: Record<string, unknown>
  defense_entry_mode?: string
  access_config?: Record<string, unknown>
  runtime_config?: Record<string, unknown>
  created_by?: ID
  last_verified_at?: ISODateTime
  created_at: ISODateTime
  updated_at: ISODateTime
}

export interface AdminAwdChallengeImportPreview {
  id: ID
  file_name: string
  slug: string
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  description: string
  service_type: AWDServiceType
  deployment_mode: AWDDeploymentMode
  version: string
  checker_type: AWDCheckerType
  checker_config?: Record<string, unknown>
  flag_mode?: string
  flag_config?: Record<string, unknown>
  defense_entry_mode?: string
  access_config?: Record<string, unknown>
  runtime_config?: Record<string, unknown>
  warnings?: string[]
  created_at: ISODateTime
}

export interface AdminAwdChallengeImportCommitData {
  challenge: AdminAwdChallengeData
}

export interface AdminChallengeListItem {
  id: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  status: ChallengeStatus
  points: number
  instance_sharing?: InstanceSharing
  created_at: ISODateTime
  updated_at?: ISODateTime
  description?: string
  created_by?: ID
  image_id?: string
  attachment_url?: string
  hints?: AdminChallengeHint[]
  flag_config?: {
    configured: boolean
    flag_type?: FlagType
    flag_regex?: string
    flag_prefix?: string
  }
}

export type AdminChallengePublishRequestStatus = 'queued' | 'running' | 'succeeded' | 'failed'

export interface ChallengeSelfCheckStepData {
  name: string
  passed: boolean
  message: string
}

export interface ChallengeSelfCheckPhaseData {
  passed: boolean
  started_at: ISODateTime
  ended_at: ISODateTime
  steps: ChallengeSelfCheckStepData[]
}

export interface ChallengeSelfCheckRuntimeData extends ChallengeSelfCheckPhaseData {
  access_url?: string
  container_count: number
  network_count: number
}

export interface ChallengeSelfCheckData {
  challenge_id: ID
  precheck: ChallengeSelfCheckPhaseData
  runtime: ChallengeSelfCheckRuntimeData
}

export interface AdminChallengePublishRequestData {
  id: ID
  challenge_id: ID
  status: AdminChallengePublishRequestStatus
  active: boolean
  requested_by?: ID
  request_source?: string
  failure_summary?: string
  started_at?: ISODateTime
  finished_at?: ISODateTime
  published_at?: ISODateTime
  result?: ChallengeSelfCheckData
  created_at: ISODateTime
  updated_at: ISODateTime
}

export interface AdminChallengeUpsertData {
  challenge: AdminChallengeListItem
}

export interface AdminChallengeImportCommitData {
  challenge: AdminChallengeListItem
}

export type ImageSourceType = 'registry' | 'dockerfile' | 'upload'
export type ImageStatus = 'pending' | 'building' | 'available' | 'failed'

export interface AdminImageListItem {
  id: ID
  name: string
  tag: string
  status: ImageStatus
  description?: string
  size_bytes?: number
  created_at: ISODateTime
  updated_at?: ISODateTime
}

export interface AdminImageCreateData {
  image: AdminImageListItem
}

export interface AuditLogItem {
  id: ID
  action: string
  resource_type: string
  resource_id?: ID
  actor_user_id?: ID
  actor_username: string
  ip?: string
  user_agent?: string
  created_at: ISODateTime
  detail?: Record<string, unknown>
}

export interface WsTicketData {
  ticket: string
  expires_at: ISODateTime
}

export interface WsMessage<T> {
  type: string
  payload: T
  timestamp: ISODateTime
}
