import { request } from './request'

import type {
  AdminChallengeListItem,
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

type AdminContestStatus = Extract<ContestStatus, 'draft' | 'registering' | 'running' | 'frozen' | 'ended'>
type AdminContestMode = Extract<ContestMode, 'jeopardy' | 'awd'>

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

export async function getDashboard(): Promise<AdminDashboardData> {
  return request<AdminDashboardData>({ method: 'GET', url: '/admin/dashboard' })
}

export async function getUsers(params?: Record<string, unknown>) {
  return request<PageResult<AdminUserListItem>>({ method: 'GET', url: '/admin/users', params })
}

export async function createUser(data: Record<string, unknown>) {
  return request<AdminUserUpsertData>({ method: 'POST', url: '/admin/users', data })
}

export async function updateUser(id: string, data: Record<string, unknown>) {
  return request<AdminUserUpsertData>({ method: 'PUT', url: `/admin/users/${encodeURIComponent(id)}`, data })
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
  return request<PageResult<AdminChallengeListItem>>({ method: 'GET', url: '/admin/challenges', params })
}

export async function getChallengeDetail(id: string) {
  return request<AdminChallengeListItem>({ method: 'GET', url: `/admin/challenges/${encodeURIComponent(id)}` })
}

export async function createChallenge(data: Record<string, unknown>) {
  return request<AdminChallengeUpsertData>({ method: 'POST', url: '/admin/challenges', data })
}

export async function updateChallenge(id: string, data: Record<string, unknown>) {
  return request<AdminChallengeUpsertData>({ method: 'PUT', url: `/admin/challenges/${encodeURIComponent(id)}`, data })
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

export async function getContests(params?: ContestListParams): Promise<PageResult<ContestDetailData>> {
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

export async function createContest(data: AdminContestCreatePayload): Promise<{ contest: ContestDetailData }> {
  const contest = await request<RawContestItem>({
    method: 'POST',
    url: '/admin/contests',
    data: serializeContestPayload(data),
  })

  return { contest: normalizeContest(contest) }
}

export async function updateContest(id: string, data: AdminContestUpdatePayload): Promise<{ contest: ContestDetailData }> {
  const contest = await request<RawContestItem>({
    method: 'PUT',
    url: `/admin/contests/${encodeURIComponent(id)}`,
    data: serializeContestPayload(data),
  })

  return { contest: normalizeContest(contest) }
}
