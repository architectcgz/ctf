import { request } from './request'

import type {
  AdminChallengeListItem,
  AdminCheatDetectionData,
  AdminChallengeUpsertData,
  AdminDashboardData,
  AdminImageCreateData,
  AdminImageListItem,
  AdminUserImportData,
  AdminUserListItem,
  AdminUserUpsertData,
  AuditLogItem,
  ContestMode,
  ContestDetailData,
  ContestListItem,
  ContestStatus,
  PageResult,
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
  password: string
  email?: string
  student_no?: string
  teacher_no?: string
  class_name?: string
  role: UserRole
  status?: UserStatus
}

export interface AdminUserUpdatePayload {
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

interface RawAdminUser {
  id: string | number
  username: string
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

function normalizeContestStatus(status: RawContestItem['status']): AdminContestStatus {
  if (status === 'registration') {
    return 'registering'
  }
  return status
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
      size: params?.page_size,
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
  return request<PageResult<AdminChallengeListItem>>({
    method: 'GET',
    url: '/admin/challenges',
    params,
  })
}

export async function getChallengeDetail(id: string) {
  return request<AdminChallengeListItem>({
    method: 'GET',
    url: `/admin/challenges/${encodeURIComponent(id)}`,
  })
}

export async function createChallenge(data: Record<string, unknown>) {
  return request<AdminChallengeUpsertData>({ method: 'POST', url: '/admin/challenges', data })
}

export async function updateChallenge(id: string, data: Record<string, unknown>) {
  return request<AdminChallengeUpsertData>({
    method: 'PUT',
    url: `/admin/challenges/${encodeURIComponent(id)}`,
    data,
  })
}

export async function deleteChallenge(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/challenges/${encodeURIComponent(id)}` })
}

export async function getImages(params?: Record<string, unknown>) {
  return request<PageResult<AdminImageListItem>>({ method: 'GET', url: '/admin/images', params })
}

export async function createImage(data: Record<string, unknown>) {
  return request<AdminImageCreateData>({ method: 'POST', url: '/admin/images', data })
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
      size: params?.page_size,
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
