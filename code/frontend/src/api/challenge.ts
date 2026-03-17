import { ApiError, request } from './request'
import { normalizeInstanceData } from './instance'

import type {
  ChallengeDetailData,
  ChallengeListItem,
  ChallengeWriteupData,
  InstanceData,
  PageResult,
  SubmitFlagData,
  UnlockHintData,
} from './contracts'

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

interface RawChallengeWriteupData extends Omit<ChallengeWriteupData, 'id' | 'challenge_id'> {
  id: string | number
  challenge_id: string | number
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
    need_target: item.need_target ?? true,
  }
}

function normalizeChallengeWriteup(item: RawChallengeWriteupData): ChallengeWriteupData {
  return {
    ...item,
    id: String(item.id),
    challenge_id: String(item.challenge_id),
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

export async function getChallengeWriteup(id: string): Promise<ChallengeWriteupData | null> {
  try {
    const payload = await request<RawChallengeWriteupData>({
      method: 'GET',
      url: `/challenges/${encodeURIComponent(id)}/writeup`,
      suppressErrorToast: true,
    })
    return normalizeChallengeWriteup(payload)
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

export async function submitFlag(id: string, flag: string): Promise<SubmitFlagData> {
  return request<SubmitFlagData>({ method: 'POST', url: `/challenges/${encodeURIComponent(id)}/submit`, data: { flag } })
}

export async function unlockHint(id: string, level: number): Promise<UnlockHintData> {
  return request<UnlockHintData>({ method: 'POST', url: `/challenges/${encodeURIComponent(id)}/hints/${level}/unlock` })
}

export async function createInstance(id: string): Promise<InstanceData> {
  const payload = await request<InstanceData & { id: string | number; challenge_id: string | number }>({
    method: 'POST',
    url: `/challenges/${encodeURIComponent(id)}/instances`,
    suppressErrorToast: true,
  })
  return normalizeInstanceData(payload)
}
