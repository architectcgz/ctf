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

export type GetContestsData = PageResult<ContestListItem>

export async function getContests(params?: Record<string, unknown>): Promise<GetContestsData> {
  return request<GetContestsData>({ method: 'GET', url: '/contests', params })
}

export async function getContestDetail(id: string): Promise<ContestDetailData> {
  return request<ContestDetailData>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}` })
}

export async function registerContest(id: string) {
  return request<void>({ method: 'POST', url: `/contests/${encodeURIComponent(id)}/register` })
}

export async function getContestChallenges(id: string): Promise<ContestChallengeItem[]> {
  return request<ContestChallengeItem[]>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}/challenges` })
}

export async function submitContestFlag(contestId: string, contestChallengeId: string, flag: string): Promise<SubmitFlagData> {
  return request<SubmitFlagData>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/challenges/${encodeURIComponent(contestChallengeId)}/submissions`,
    data: { flag },
  })
}

export async function getScoreboard(id: string, params?: Record<string, unknown>): Promise<ContestScoreboardData> {
  return request<ContestScoreboardData>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}/scoreboard`, params })
}

export async function getAnnouncements(id: string): Promise<ContestAnnouncement[]> {
  return request<ContestAnnouncement[]>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}/announcements` })
}

export async function createTeam(id: string, data: Record<string, unknown>): Promise<TeamData> {
  return request<TeamData>({ method: 'POST', url: `/contests/${encodeURIComponent(id)}/teams`, data })
}

export async function joinTeam(contestId: string, teamId: string) {
  return request<void>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/teams/${encodeURIComponent(teamId)}/join`,
  })
}

export async function getMyProgress(id: string): Promise<ContestMyProgressData> {
  return request<ContestMyProgressData>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}/my-progress` })
}

export async function getMyTeam(contestId: string): Promise<TeamData | null> {
  return request<TeamData | null>({ method: 'GET', url: `/contests/${encodeURIComponent(contestId)}/my-team` })
}

export async function kickTeamMember(contestId: string, teamId: string, userId: string) {
  return request<void>({
    method: 'DELETE',
    url: `/contests/${encodeURIComponent(contestId)}/teams/${encodeURIComponent(teamId)}/members/${encodeURIComponent(userId)}`,
  })
}
