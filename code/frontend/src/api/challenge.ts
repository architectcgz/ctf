import { ApiError, getAxiosInstance, request } from './request'
import { normalizeInstanceData } from './instance'

import type {
  ChallengeSubmissionRecordData,
  CommunityChallengeSolutionData,
  ChallengeDetailData,
  ChallengeListItem,
  ChallengeWriteupData,
  InstanceData,
  PageResult,
  RecommendedChallengeSolutionData,
  SubmissionWriteupData,
  SubmitFlagData,
} from './contracts'

export type GetChallengesData = PageResult<ChallengeListItem>
export interface DownloadedAttachment {
  blob: Blob
  filename: string
}

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

interface RawSubmissionWriteupData extends Omit<
  SubmissionWriteupData,
  'id' | 'user_id' | 'challenge_id' | 'contest_id' | 'recommended_by'
> {
  id: string | number
  user_id: string | number
  challenge_id: string | number
  contest_id?: string | number
  recommended_by?: string | number
}

interface RawChallengeSubmissionRecordData extends Omit<ChallengeSubmissionRecordData, 'id'> {
  id: string | number
}

interface RawRecommendedChallengeSolutionData extends Omit<
  RecommendedChallengeSolutionData,
  'id' | 'source_id' | 'challenge_id'
> {
  id: string | number
  source_id: string | number
  challenge_id: string | number
}

interface RawCommunityChallengeSolutionData extends Omit<
  CommunityChallengeSolutionData,
  'id' | 'challenge_id' | 'user_id'
> {
  id: string | number
  challenge_id: string | number
  user_id: string | number
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
    instance_sharing: item.instance_sharing ?? 'per_user',
  }
}

function normalizeChallengeWriteup(item: RawChallengeWriteupData): ChallengeWriteupData {
  return {
    ...item,
    id: String(item.id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeSubmissionWriteup(item: RawSubmissionWriteupData): SubmissionWriteupData {
  return {
    ...item,
    id: String(item.id),
    user_id: String(item.user_id),
    challenge_id: String(item.challenge_id),
    contest_id: item.contest_id !== undefined ? String(item.contest_id) : undefined,
    recommended_by: item.recommended_by !== undefined ? String(item.recommended_by) : undefined,
  }
}

function normalizeChallengeSubmissionRecord(
  item: RawChallengeSubmissionRecordData
): ChallengeSubmissionRecordData {
  return {
    ...item,
    id: String(item.id),
  }
}

function normalizeRecommendedSolution(
  item: RawRecommendedChallengeSolutionData
): RecommendedChallengeSolutionData {
  return {
    ...item,
    id: String(item.id),
    source_id: String(item.source_id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeCommunitySolution(
  item: RawCommunityChallengeSolutionData
): CommunityChallengeSolutionData {
  return {
    ...item,
    id: String(item.id),
    challenge_id: String(item.challenge_id),
    user_id: String(item.user_id),
  }
}

export async function getChallenges(params?: Record<string, unknown>): Promise<GetChallengesData> {
  const payload = await request<PageResult<RawChallengeListItem>>({
    method: 'GET',
    url: '/challenges',
    params,
  })
  return {
    ...payload,
    list: payload.list.map(normalizeChallengeListItem),
  }
}

export async function getChallengeDetail(id: string): Promise<ChallengeDetailData> {
  const payload = await request<RawChallengeDetailData>({
    method: 'GET',
    url: `/challenges/${encodeURIComponent(id)}`,
  })
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

export async function getMyChallengeWriteupSubmission(
  id: string
): Promise<SubmissionWriteupData | null> {
  try {
    const payload = await request<RawSubmissionWriteupData | null>({
      method: 'GET',
      url: `/challenges/${encodeURIComponent(id)}/writeup-submissions/me`,
      suppressErrorToast: true,
    })
    if (!payload) {
      return null
    }
    return normalizeSubmissionWriteup(payload)
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

export async function getMyChallengeSubmissionRecords(
  id: string
): Promise<ChallengeSubmissionRecordData[]> {
  const payload = await request<RawChallengeSubmissionRecordData[]>({
    method: 'GET',
    url: `/challenges/${encodeURIComponent(id)}/submissions/mine`,
  })
  return payload.map(normalizeChallengeSubmissionRecord)
}

export async function upsertChallengeWriteupSubmission(
  id: string,
  payload: {
    title: string
    content: string
    submission_status: 'draft' | 'published'
  }
): Promise<SubmissionWriteupData> {
  const response = await request<RawSubmissionWriteupData>({
    method: 'POST',
    url: `/challenges/${encodeURIComponent(id)}/writeup-submissions`,
    data: payload,
  })
  return normalizeSubmissionWriteup(response)
}

export async function getRecommendedChallengeSolutions(
  id: string
): Promise<RecommendedChallengeSolutionData[]> {
  const payload = await request<PageResult<RawRecommendedChallengeSolutionData>>({
    method: 'GET',
    url: `/challenges/${encodeURIComponent(id)}/solutions/recommended`,
  })
  return payload.list.map(normalizeRecommendedSolution)
}

export async function getCommunityChallengeSolutions(
  id: string,
  params?: { q?: string; sort?: 'newest' | 'oldest'; page?: number; page_size?: number }
): Promise<PageResult<CommunityChallengeSolutionData>> {
  const payload = await request<PageResult<RawCommunityChallengeSolutionData>>({
    method: 'GET',
    url: `/challenges/${encodeURIComponent(id)}/solutions/community`,
    params,
  })
  return {
    ...payload,
    list: payload.list.map(normalizeCommunitySolution),
  }
}

export async function submitFlag(id: string, flag: string): Promise<SubmitFlagData> {
  return request<SubmitFlagData>({
    method: 'POST',
    url: `/challenges/${encodeURIComponent(id)}/submit`,
    data: { flag },
  })
}

export async function createInstance(id: string): Promise<InstanceData> {
  const payload = await request<
    InstanceData & { id: string | number; challenge_id: string | number }
  >({
    method: 'POST',
    url: `/challenges/${encodeURIComponent(id)}/instances`,
    suppressErrorToast: true,
  })
  return normalizeInstanceData(payload)
}

function resolveFilename(contentDisposition: string | undefined, fallback: string): string {
  if (!contentDisposition) return fallback

  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match?.[1]) {
    return decodeURIComponent(utf8Match[1])
  }

  const basicMatch = contentDisposition.match(/filename=\"?([^\";]+)\"?/i)
  if (basicMatch?.[1]) {
    return basicMatch[1]
  }

  return fallback
}

export async function downloadAttachment(attachmentURL: string): Promise<DownloadedAttachment> {
  const normalizedURL = normalizeAttachmentRequestURL(attachmentURL)
  const response = await getAxiosInstance().get<Blob>(normalizedURL, {
    responseType: 'blob',
  })
  const fallback = decodeURIComponent(attachmentURL.split('/').pop() || 'attachment')
  return {
    blob: response.data,
    filename: resolveFilename(response.headers['content-disposition'], fallback),
  }
}

function normalizeAttachmentRequestURL(rawURL: string): string {
  const value = rawURL.trim()
  if (!value) return value
  if (/^https?:\/\//i.test(value)) return value

  // axios instance already has baseURL=/api/v1, avoid /api/v1/api/v1/...
  if (value.startsWith('/api/v1/')) {
    return value.slice('/api/v1'.length)
  }
  return value
}
