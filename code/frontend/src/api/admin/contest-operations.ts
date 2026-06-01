import { request } from '../request'

import type {
  AdminAWDRoundCreatePayload,
  AdminAWDCurrentRoundCheckPayload,
  AdminAWDServiceCheckPayload,
  AdminAWDAttackLogPayload,
  RawAWDRoundItem,
  RawAWDTeamServiceItem,
  RawAWDAttackLogItem,
  RawAWDRoundSummaryData,
  RawAWDCheckerRunData,
  RawAWDReadinessData,
  AWDRoundData,
  AWDTeamServiceData,
  AWDAttackLogData,
  AWDRoundSummaryData,
  AWDCheckerRunData,
  AWDReadinessData,
} from './contest-support'

// Re-export types for barrel consumers
export type {
  AdminAWDRoundCreatePayload,
  AdminAWDCurrentRoundCheckPayload,
  AdminAWDServiceCheckPayload,
  AdminAWDAttackLogPayload,
} from './contest-support'

import {
  normalizeAWDRound,
  normalizeAWDTeamService,
  normalizeAWDAttackLog,
  normalizeAWDRoundSummary,
  normalizeAWDCheckerRun,
  normalizeAWDReadiness,
} from './contest-support'

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

export async function getContestAWDReadiness(
  contestId: string
): Promise<AWDReadinessData> {
  const response = await request<RawAWDReadinessData>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/readiness`,
  })
  return normalizeAWDReadiness(response)
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

export async function runContestAWDCurrentRoundCheck(
  contestId: string,
  data?: AdminAWDCurrentRoundCheckPayload
): Promise<AWDCheckerRunData> {
  const response = await request<RawAWDCheckerRunData>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/awd/current-round/check`,
    data,
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
