import { request } from './request'

import type { MyProgressData, RecommendationItem, ReportExportData, SkillProfileData, TimelineEvent } from './contracts'

export async function getSkillProfile(): Promise<SkillProfileData> {
  return request<SkillProfileData>({ method: 'GET', url: '/users/me/skill-profile' })
}

export async function getRecommendations(): Promise<RecommendationItem[]> {
  return request<RecommendationItem[]>({ method: 'GET', url: '/users/me/recommendations' })
}

export async function getMyProgress(): Promise<MyProgressData> {
  return request<MyProgressData>({ method: 'GET', url: '/users/me/progress' })
}

export async function getMyTimeline(): Promise<TimelineEvent[]> {
  return request<TimelineEvent[]>({ method: 'GET', url: '/users/me/timeline' })
}

export async function exportPersonalReport(data: Record<string, unknown>): Promise<ReportExportData> {
  return request<ReportExportData>({ method: 'POST', url: '/reports/personal', data })
}
