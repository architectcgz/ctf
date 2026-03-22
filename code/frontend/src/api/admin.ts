import { ApiError, request } from './request'

import type {
  AWDAttackLogData,
  AWDCheckerRunData,
  AWDRoundData,
  AWDRoundMetricsData,
  AWDRoundSummaryData,
  AWDRoundSummaryItemData,
  AWDTeamServiceData,
  AdminContestChallengeData,
  AdminContestTeamData,
  AdminChallengeHint,
  AdminChallengeListItem,
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
  ContestListItem,
  ContestScoreboardData,
  ContestStatus,
  EnvironmentTemplateData,
  PageResult,
  TopologyLinkData,
  TopologyNetworkData,
  TopologyNodeData,
  TopologyTrafficPolicyData,
  WriteupVisibility,
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
  check_result?: Record<string, unknown> | null
  attack_received: number
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

interface RawAWDRoundSummaryItem {
  team_id: string | number
  team_name: string
  service_up_count: number
  service_down_count: number
  service_compromised_count: number
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

interface RawAWDCheckerRunData {
  round: RawAWDRoundItem
  services: RawAWDTeamServiceItem[]
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
  created_at: string
}

interface RawAdminChallengeItem {
  id: string | number
  title: string
  description?: string
  category: AdminChallengeListItem['category']
  difficulty: AdminChallengeListItem['difficulty']
  points: number
  image_id?: string | number | null
  attachment_url?: string
  hints?: Array<{
    id: string | number
    level: number
    title?: string
    cost_points?: number
    content: string
  }>
  status: 'draft' | 'published' | 'archived'
  created_at: string
  updated_at: string
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
  flag_type: 'static' | 'dynamic'
  flag_prefix?: string
  configured: boolean
}

interface RawAdminChallengeWriteupData {
  id: string | number
  challenge_id: string | number
  title: string
  content: string
  visibility: WriteupVisibility
  release_at?: string | null
  created_by?: string | number | null
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
}

export interface AdminAWDRoundCreatePayload {
  round_number: number
  status?: AWDRoundData['status']
  attack_score?: number
  defense_score?: number
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
    release_at: item.release_at || undefined,
    created_by: item.created_by == null ? undefined : String(item.created_by),
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

function normalizeAWDTeamService(item: RawAWDTeamServiceItem): AWDTeamServiceData {
  return {
    id: String(item.id),
    round_id: String(item.round_id),
    team_id: String(item.team_id),
    team_name: item.team_name,
    challenge_id: String(item.challenge_id),
    service_status: item.service_status,
    check_result: item.check_result || {},
    attack_received: item.attack_received,
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
    cost_points: hint.cost_points,
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
    image_id: item.image_id == null ? undefined : String(item.image_id),
    attachment_url: item.attachment_url,
    hints,
    created_at: item.created_at,
    updated_at: item.updated_at,
    flag_config: flagConfig
      ? {
          configured: flagConfig.configured,
          flag_type: flagConfig.flag_type,
          flag_prefix: flagConfig.flag_prefix,
        }
      : undefined,
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
    url: '/admin/challenges',
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
      url: `/admin/challenges/${encodeURIComponent(id)}`,
    }),
    request<RawChallengeFlagConfig>({
      method: 'GET',
      url: `/admin/challenges/${encodeURIComponent(id)}/flag`,
    }).catch(() => undefined),
  ])

  return normalizeChallenge(challenge, flagConfig)
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
  flag_type: 'static' | 'dynamic'
  flag?: string
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
    url: '/admin/challenges',
    data,
  })
  return {
    challenge: normalizeChallenge(response),
  }
}

export async function updateChallenge(id: string, data: Partial<AdminChallengePayload>) {
  await request<void>({
    method: 'PUT',
    url: `/admin/challenges/${encodeURIComponent(id)}`,
    data,
  })
}

export async function configureChallengeFlag(id: string, data: AdminChallengeFlagPayload) {
  return request<{ message: string }>({
    method: 'PUT',
    url: `/admin/challenges/${encodeURIComponent(id)}/flag`,
    data,
  })
}

export async function getChallengeFlagConfig(id: string) {
  return request<RawChallengeFlagConfig>({
    method: 'GET',
    url: `/admin/challenges/${encodeURIComponent(id)}/flag`,
  })
}

export async function publishChallenge(id: string) {
  return request<void>({
    method: 'PUT',
    url: `/admin/challenges/${encodeURIComponent(id)}/publish`,
  })
}

export async function deleteChallenge(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/challenges/${encodeURIComponent(id)}` })
}

export interface AdminChallengeWriteupPayload {
  title: string
  content: string
  visibility: WriteupVisibility
  release_at?: string
}

export async function getChallengeWriteup(id: string): Promise<AdminChallengeWriteupData | null> {
  try {
    const response = await request<RawAdminChallengeWriteupData>({
      method: 'GET',
      url: `/admin/challenges/${encodeURIComponent(id)}/writeup`,
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
    url: `/admin/challenges/${encodeURIComponent(id)}/writeup`,
    data,
  })
  return normalizeAdminChallengeWriteup(response)
}

export async function deleteChallengeWriteup(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/admin/challenges/${encodeURIComponent(id)}/writeup`,
  })
}

export async function getChallengeTopology(id: string): Promise<ChallengeTopologyData | null> {
  try {
    const response = await request<RawChallengeTopologyData>({
      method: 'GET',
      url: `/admin/challenges/${encodeURIComponent(id)}/topology`,
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
    url: `/admin/challenges/${encodeURIComponent(id)}/topology`,
    data,
  })
  return normalizeChallengeTopology(response)
}

export async function deleteChallengeTopology(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/admin/challenges/${encodeURIComponent(id)}/topology`,
  })
}

export async function getEnvironmentTemplates(
  keyword?: string
): Promise<EnvironmentTemplateData[]> {
  const response = await request<RawEnvironmentTemplateData[]>({
    method: 'GET',
    url: '/admin/environment-templates',
    params: { keyword },
  })
  return response.map(normalizeEnvironmentTemplate)
}

export async function getEnvironmentTemplate(id: string): Promise<EnvironmentTemplateData> {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'GET',
    url: `/admin/environment-templates/${encodeURIComponent(id)}`,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function createEnvironmentTemplate(data: AdminEnvironmentTemplatePayload) {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'POST',
    url: '/admin/environment-templates',
    data,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function updateEnvironmentTemplate(id: string, data: AdminEnvironmentTemplatePayload) {
  const response = await request<RawEnvironmentTemplateData>({
    method: 'PUT',
    url: `/admin/environment-templates/${encodeURIComponent(id)}`,
    data,
  })
  return normalizeEnvironmentTemplate(response)
}

export async function deleteEnvironmentTemplate(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/admin/environment-templates/${encodeURIComponent(id)}`,
  })
}

export async function getImages(params?: Record<string, unknown>) {
  const response = await request<PageResult<RawImageItem>>({
    method: 'GET',
    url: '/admin/images',
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
  const response = await request<RawImageItem>({ method: 'POST', url: '/admin/images', data })
  return {
    image: normalizeImage(response),
  }
}

export async function deleteImage(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/images/${encodeURIComponent(id)}` })
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
  data: AdminContestUpdatePayload
): Promise<{ contest: ContestDetailData }> {
  const contest = await request<RawContestItem>({
    method: 'PUT',
    url: `/admin/contests/${encodeURIComponent(id)}`,
    data: serializeContestPayload(data),
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
  })
  return normalizeAWDRound(response)
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

export async function runContestAWDCurrentRoundCheck(contestId: string): Promise<AWDCheckerRunData> {
  const response = await request<RawAWDCheckerRunData>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/current-round/check`,
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
