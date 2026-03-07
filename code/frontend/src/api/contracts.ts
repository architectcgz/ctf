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
export type ChallengeDifficulty = 'beginner' | 'easy' | 'medium' | 'hard' | 'hell'

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
  cost_points?: number
  is_unlocked: boolean
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
  attachment_url?: string
  is_solved: boolean
  solved_at?: ISODateTime
  hints: ChallengeHint[]
}

export type InstanceStatus = 'pending' | 'creating' | 'running' | 'expired' | 'destroying' | 'destroyed' | 'failed' | 'crashed'
export type FlagType = 'static' | 'dynamic' | 'regex'

export interface InstanceData {
  id: ID
  challenge_id: ID
  status: InstanceStatus
  access_url?: string
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

export interface SubmitFlagSuccessData {
  correct: true
  points_earned: number
  first_blood: boolean
  solved_at: ISODateTime
  challenge_progress: { total_challenges: number; solved_challenges: number }
}

export interface SubmitFlagFailureData {
  correct: false
  remaining_attempts?: number
}

export type SubmitFlagData = SubmitFlagSuccessData | SubmitFlagFailureData

export interface UnlockHintData {
  hint: ChallengeHint
  points_spent?: number
  remaining_points?: number
}

export interface MyProgressData {
  total_challenges: number
  solved_challenges: number
  by_category?: Record<string, { total: number; solved: number }>
  by_difficulty?: Record<string, { total: number; solved: number }>
}

export interface TimelineEvent {
  id: ID
  type: 'solve' | 'submit' | 'instance' | 'hint'
  title: string
  created_at: ISODateTime
  meta?: Record<string, unknown>
}

export type ContestMode = 'jeopardy' | 'awd' | 'awd_plus' | 'king_of_hill'
export type ContestStatus = 'draft' | 'published' | 'registering' | 'running' | 'frozen' | 'ended' | 'cancelled' | 'archived'

export interface ContestListItem {
  id: ID
  title: string
  mode: ContestMode
  status: ContestStatus
  starts_at: ISODateTime
  ends_at: ISODateTime
  register_ends_at?: ISODateTime
}

export interface ContestDetailData extends ContestListItem {
  description?: string
  rules?: string
  team_size_limit?: number
  scoreboard_frozen?: boolean
}

export interface ContestChallengeItem {
  id: ID
  challenge_id: ID
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

export interface ContestScoreboardData {
  contest: { id: ID; title: string; status: ContestStatus; started_at: ISODateTime; ends_at: ISODateTime }
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

export type NotificationType = 'system' | 'contest' | 'challenge' | 'team'

export interface NotificationItem {
  id: ID
  type: NotificationType
  title: string
  content?: string
  level?: 'info' | 'success' | 'warning' | 'error'
  unread: boolean
  created_at: ISODateTime
}

export interface TeacherClassItem {
  name: string
  student_count?: number
}

export interface TeacherStudentItem {
  id: ID
  username: string
  name?: string
  progress?: MyProgressData
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
  status: 'ready' | 'processing'
  download_url?: string
  expires_at?: ISODateTime
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

export type ChallengeStatus = 'draft' | 'review' | 'active' | 'archived'

export interface AdminChallengeListItem {
  id: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  status: ChallengeStatus
  base_score: number
  solve_count: number
  created_at: ISODateTime
  description?: string
  hints?: string[]
  image_id?: string
  image_name?: string
  flag?: string
  resource_limits?: { cpu: number; memory: number }
  tags?: string[]
}

export interface AdminChallengeUpsertData {
  challenge: AdminChallengeListItem
}

export type ImageSourceType = 'registry' | 'dockerfile' | 'upload'
export type ImageStatus = 'pending' | 'building' | 'ready' | 'failed' | 'deprecated'

export interface AdminImageListItem {
  id: ID
  name: string
  tag: string
  source_type: ImageSourceType
  status: ImageStatus
  size_bytes?: number
  created_at: ISODateTime
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
