import { request } from '../request'

import type { ReportExportData } from '../contracts'
import type {
  AdminContestStatus,
  AdminContestMode,
  AdminContestCreatePayload,
  AdminContestUpdatePayload,
  AdminContestChallengeCreatePayload,
  AdminContestChallengeUpdatePayload,
  RawContestItem,
  RawContestPageResult,
  RawAdminContestTeamItem,
  RawAdminContestChallengeItem,
  ContestDetailData,
  ContestPageData,
  ContestScoreboardData,
  AdminContestTeamData,
  AdminContestChallengeRelationData,
} from './contest-support'

// Re-export types for barrel consumers
export type {
  AdminContestCreatePayload,
  AdminContestUpdatePayload,
  AdminContestChallengeCreatePayload,
  AdminContestChallengeUpdatePayload,
  AdminContestStatus,
  AdminContestMode,
  ContestDetailData,
  ContestPageData,
  ContestScoreboardData,
  AdminContestTeamData,
  AdminContestChallengeRelationData,
} from './contest-support'

import {
  normalizeContest,
  normalizeContestSummary,
  normalizeAdminContestTeam,
  normalizeAdminContestChallenge,
  normalizeContestScoreboard,
  serializeContestPayload,
  serializeContestStatus,
  serializeContestStatuses,
} from './contest-support'

/* ── 内部类型 ── */

interface ContestListParams {
  page?: number
  page_size?: number
  status?: AdminContestStatus
  statuses?: AdminContestStatus[]
  mode?: AdminContestMode
  sort_key?: 'created_at' | 'start_time'
  sort_order?: 'asc' | 'desc'
}

/* ── API 函数 ── */

export async function getContests(
  params?: ContestListParams,
  options?: { signal?: AbortSignal }
): Promise<ContestPageData<ContestDetailData>> {
  const response = await request<RawContestPageResult<RawContestItem>>({
    method: 'GET',
    url: '/admin/contests',
    params: {
      page: params?.page,
      page_size: params?.page_size,
      status: serializeContestStatus(params?.status),
      statuses: serializeContestStatuses(params?.statuses),
      mode: params?.mode,
      sort_key: params?.sort_key,
      sort_order: params?.sort_order,
    },
    signal: options?.signal,
  })

  return {
    ...response,
    list: response.list.map(normalizeContest),
    summary: normalizeContestSummary(response.summary),
  }
}

export async function getContest(id: string): Promise<ContestDetailData> {
  const contest = await request<RawContestItem>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(id)}`,
  })

  return normalizeContest(contest)
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

export async function listContestTeams(
  contestId: string
): Promise<AdminContestTeamData[]> {
  const response = await request<RawAdminContestTeamItem[]>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(contestId)}/teams`,
  })
  return response.map(normalizeAdminContestTeam)
}

export async function listAdminContestChallenges(
  contestId: string
): Promise<AdminContestChallengeRelationData[]> {
  const response = await request<RawAdminContestChallengeItem[]>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/challenges`,
  })
  return response.map(normalizeAdminContestChallenge)
}

export async function createAdminContestChallenge(
  contestId: string,
  data: AdminContestChallengeCreatePayload
): Promise<AdminContestChallengeRelationData> {
  const response = await request<RawAdminContestChallengeItem>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/challenges`,
    data: {
      challenge_id: data.challenge_id,
      points: data.points,
      order: data.order,
      is_visible: data.is_visible,
    },
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
    data: {
      points: data.points,
      order: data.order,
      is_visible: data.is_visible,
    },
  })
}

export async function deleteAdminContestChallenge(
  contestId: string,
  challengeId: string
): Promise<void> {
  await request<void>({
    method: 'DELETE',
    url: `/admin/contests/${encodeURIComponent(contestId)}/challenges/${encodeURIComponent(challengeId)}`,
  })
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
