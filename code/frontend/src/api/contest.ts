import { request } from './request'

import type {
  ContestAnnouncement,
  ContestChallengeItem,
  ContestDetailData,
  ContestListItem,
  ContestMyProgressData,
  ContestScoreboardData,
  PageResult,
  SubmitFlagData,
  TeamData,
} from './contracts'

type RawContestStatus = 'draft' | 'registration' | 'running' | 'frozen' | 'ended'

interface RawContestItem {
  id: string | number
  title: string
  description?: string
  mode: ContestDetailData['mode']
  status: RawContestStatus
  start_time: string
  end_time: string
  freeze_time?: string | null
}

interface RawTeamMember {
  user_id: string | number
  username: string
  joined_at: string
}

interface RawTeamData {
  id: string | number
  name: string
  invite_code?: string
  captain_user_id: string | number
  members: RawTeamMember[]
}

interface RawTeamCreateResp {
  id: string | number
  name: string
  contest_id: string | number
  captain_id: string | number
  invite_code?: string
  max_members: number
  member_count: number
  created_at: string
}

export type GetContestsData = PageResult<ContestListItem>

function normalizeContestStatus(status: RawContestStatus): ContestListItem['status'] {
  if (status === 'registration') {
    return 'registering'
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

function normalizeTeam(payload: RawTeamData | null): TeamData | null {
  if (!payload) {
    return null
  }

  return {
    id: String(payload.id),
    name: payload.name,
    invite_code: payload.invite_code,
    captain_user_id: String(payload.captain_user_id),
    members: payload.members.map((member) => ({
      user_id: String(member.user_id),
      username: member.username,
      joined_at: member.joined_at,
    })),
  }
}

export async function getContests(params?: Record<string, unknown>): Promise<GetContestsData> {
  const response = await request<PageResult<RawContestItem>>({
    method: 'GET',
    url: '/contests',
    params,
  })
  return {
    ...response,
    list: response.list.map(normalizeContest),
  }
}

export async function getContestDetail(id: string): Promise<ContestDetailData> {
  const response = await request<RawContestItem>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}`,
  })
  return normalizeContest(response)
}

export async function registerContest(id: string) {
  return request<void>({ method: 'POST', url: `/contests/${encodeURIComponent(id)}/register` })
}

export async function getContestChallenges(id: string): Promise<ContestChallengeItem[]> {
  return request<ContestChallengeItem[]>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}/challenges`,
  })
}

export async function submitContestFlag(
  contestId: string,
  contestChallengeId: string,
  flag: string
): Promise<SubmitFlagData> {
  return request<SubmitFlagData>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/challenges/${encodeURIComponent(contestChallengeId)}/submissions`,
    data: { flag },
  })
}

export async function getScoreboard(
  id: string,
  params?: Record<string, unknown>
): Promise<ContestScoreboardData> {
  return request<ContestScoreboardData>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}/scoreboard`,
    params,
  })
}

export async function getAnnouncements(id: string): Promise<ContestAnnouncement[]> {
  return request<ContestAnnouncement[]>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}/announcements`,
  })
}

export async function createTeam(id: string, data: Record<string, unknown>): Promise<void> {
  await request<RawTeamCreateResp>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(id)}/teams`,
    data,
  })
}

export async function joinTeam(contestId: string, teamId: string) {
  return request<void>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/teams/${encodeURIComponent(teamId)}/join`,
  })
}

export async function getMyProgress(id: string): Promise<ContestMyProgressData> {
  return request<ContestMyProgressData>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(id)}/my-progress`,
  })
}

export async function getMyTeam(contestId: string): Promise<TeamData | null> {
  const response = await request<RawTeamData | null>({
    method: 'GET',
    url: `/contests/${encodeURIComponent(contestId)}/my-team`,
  })
  return normalizeTeam(response)
}

export async function kickTeamMember(contestId: string, teamId: string, userId: string) {
  return request<void>({
    method: 'DELETE',
    url: `/contests/${encodeURIComponent(contestId)}/teams/${encodeURIComponent(teamId)}/members/${encodeURIComponent(userId)}`,
  })
}
