import { request } from './request'

export async function getContests(params?: Record<string, unknown>) {
  return request<{ items: unknown[]; total: number }>({ method: 'GET', url: '/contests', params })
}

export async function getContestDetail(id: string) {
  return request<unknown>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}` })
}

export async function registerContest(id: string) {
  return request<void>({ method: 'POST', url: `/contests/${encodeURIComponent(id)}/register` })
}

export async function getContestChallenges(id: string) {
  return request<unknown[]>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}/challenges` })
}

export async function submitContestFlag(contestId: string, contestChallengeId: string, flag: string) {
  return request<unknown>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/challenges/${encodeURIComponent(contestChallengeId)}/submissions`,
    data: { flag },
  })
}

export async function getScoreboard(id: string, params?: Record<string, unknown>) {
  return request<unknown>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}/scoreboard`, params })
}

export async function getAnnouncements(id: string) {
  return request<unknown[]>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}/announcements` })
}

export async function createTeam(id: string, data: Record<string, unknown>) {
  return request<unknown>({ method: 'POST', url: `/contests/${encodeURIComponent(id)}/teams`, data })
}

export async function joinTeam(contestId: string, teamId: string, code: string) {
  return request<void>({
    method: 'POST',
    url: `/contests/${encodeURIComponent(contestId)}/teams/${encodeURIComponent(teamId)}/join`,
    data: { code },
  })
}

export async function getMyProgress(id: string) {
  return request<unknown>({ method: 'GET', url: `/contests/${encodeURIComponent(id)}/my-progress` })
}

