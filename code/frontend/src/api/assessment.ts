import { getAxiosInstance, request } from './request'

import type { MyProgressData, RecommendationItem, ReportExportData, SkillProfileData, TimelineEvent } from './contracts'
import {
  normalizeRecommendations,
  normalizeSkillProfile,
  type RawRecommendationResponse,
  type RawSkillProfileResponse,
} from '@/utils/skillProfile'

interface RawProgressResponse {
  total_score: number
  total_solved: number
  rank: number
  category_stats?: Array<{ category: string; solved: number; total: number }>
  difficulty_stats?: Array<{ difficulty: string; solved: number; total: number }>
}

interface RawTimelineItem {
  type: string
  challenge_id: string | number
  title: string
  timestamp: string
  is_correct?: boolean
  points?: number
}

interface RawTimelineResponse {
  events: RawTimelineItem[]
}

export async function getSkillProfile(): Promise<SkillProfileData> {
  const payload = await request<RawSkillProfileResponse>({ method: 'GET', url: '/users/me/skill-profile' })
  return normalizeSkillProfile(payload)
}

export async function getRecommendations(): Promise<RecommendationItem[]> {
  const payload = await request<RawRecommendationResponse>({ method: 'GET', url: '/users/me/recommendations' })
  return normalizeRecommendations(payload)
}

export async function getMyProgress(): Promise<MyProgressData> {
  const payload = await request<RawProgressResponse>({ method: 'GET', url: '/users/me/progress' })
  return {
    total_score: payload.total_score,
    total_solved: payload.total_solved,
    rank: payload.rank,
    category_stats: payload.category_stats ?? [],
    difficulty_stats: payload.difficulty_stats ?? [],
  }
}

export async function getMyTimeline(): Promise<TimelineEvent[]> {
  const payload = await request<RawTimelineResponse>({ method: 'GET', url: '/users/me/timeline' })
  return payload.events.map((item) => ({
    id: `${item.type}-${item.challenge_id}-${item.timestamp}`,
    type: item.type === 'instance_start' || item.type === 'instance_destroy'
      ? 'instance'
      : item.type === 'flag_submit' && item.is_correct
        ? 'solve'
        : item.type === 'flag_submit'
          ? 'submit'
          : item.type,
    title: item.title,
    created_at: item.timestamp,
    challenge_id: String(item.challenge_id),
    is_correct: item.is_correct,
    points: item.points,
    meta: {
      raw_type: item.type,
    },
  }))
}

export async function exportPersonalReport(data: Record<string, unknown>): Promise<ReportExportData> {
  const payload = await request<ReportExportData & { report_id: string | number }>({ method: 'POST', url: '/reports/personal', data })
  return {
    ...payload,
    report_id: String(payload.report_id),
  }
}

export interface DownloadedReport {
  blob: Blob
  filename: string
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

export async function downloadReport(reportId: string | number): Promise<DownloadedReport> {
  const normalizedId = encodeURIComponent(String(reportId))
  const response = await getAxiosInstance().get<Blob>(`/reports/${normalizedId}/download`, {
    responseType: 'blob',
  })

  return {
    blob: response.data,
    filename: resolveFilename(response.headers['content-disposition'], `report-${normalizedId}`),
  }
}

export interface DownloadedReport {
  blob: Blob
  filename: string
}

function resolveFilename(contentDisposition: string | undefined, fallback: string): string {
  if (!contentDisposition) return fallback

  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8Match?.[1]) {
    return decodeURIComponent(utf8Match[1])
  }

  const basicMatch = contentDisposition.match(/filename="?([^";]+)"?/i)
  if (basicMatch?.[1]) {
    return basicMatch[1]
  }

  return fallback
}

export async function downloadReport(reportId: string | number): Promise<DownloadedReport> {
  const normalizedId = encodeURIComponent(String(reportId))
  const response = await getAxiosInstance().get<Blob>(`/reports/${normalizedId}/download`, {
    responseType: 'blob',
  })

  return {
    blob: response.data,
    filename: resolveFilename(
      response.headers['content-disposition'],
      `report-${normalizedId}`
    ),
  }
}
