import { request } from './request'

import type { ChallengeDetailData, ChallengeListItem, InstanceData, PageResult, SubmitFlagData, UnlockHintData } from './contracts'

export type GetChallengesData = PageResult<ChallengeListItem>

interface RawChallengeListItem extends Omit<ChallengeListItem, 'id' | 'tags'> {
  id: string | number
  tags?: string[]
}

interface RawChallengeDetailData extends Omit<ChallengeDetailData, 'id' | 'tags' | 'hints'> {
  id: string | number
  tags?: string[]
  hints?: ChallengeDetailData['hints']
}

function normalizeChallengeListItem(item: RawChallengeListItem): ChallengeListItem {
  return {
    ...item,
    id: String(item.id),
    tags: item.tags ?? [],
  }
}

function normalizeChallengeDetail(item: RawChallengeDetailData): ChallengeDetailData {
  return {
    ...item,
    id: String(item.id),
    tags: item.tags ?? [],
    hints: item.hints ?? [],
  }
}

export async function getChallenges(params?: Record<string, unknown>): Promise<GetChallengesData> {
  const payload = await request<PageResult<RawChallengeListItem>>({ method: 'GET', url: '/challenges', params })
  return {
    ...payload,
    list: payload.list.map(normalizeChallengeListItem),
  }
}

export async function getChallengeDetail(id: string): Promise<ChallengeDetailData> {
  const payload = await request<RawChallengeDetailData>({ method: 'GET', url: `/challenges/${encodeURIComponent(id)}` })
  return normalizeChallengeDetail(payload)
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
