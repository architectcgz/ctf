import { request } from './request'

import type { ChallengeDetailData, ChallengeListItem, InstanceData, PageResult, SubmitFlagData, UnlockHintData } from './contracts'

export type GetChallengesData = PageResult<ChallengeListItem>

export async function getChallenges(params?: Record<string, unknown>): Promise<GetChallengesData> {
  return request<GetChallengesData>({ method: 'GET', url: '/challenges', params })
}

export async function getChallengeDetail(id: string): Promise<ChallengeDetailData> {
  return request<ChallengeDetailData>({ method: 'GET', url: `/challenges/${encodeURIComponent(id)}` })
}

export async function submitFlag(id: string, flag: string): Promise<SubmitFlagData> {
  return request<SubmitFlagData>({ method: 'POST', url: `/challenges/${encodeURIComponent(id)}/submit`, data: { flag } })
}

export async function unlockHint(id: string, level: number): Promise<UnlockHintData> {
  return request<UnlockHintData>({ method: 'POST', url: `/challenges/${encodeURIComponent(id)}/hints/${level}/unlock` })
}

export async function createInstance(id: string): Promise<InstanceData> {
  return request<InstanceData>({ method: 'POST', url: `/challenges/${encodeURIComponent(id)}/instances` })
}
