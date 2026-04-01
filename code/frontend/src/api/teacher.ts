import { request } from './request'

import type {
  TeacherEvidenceData,
  MyProgressData,
  RecommendationItem,
  ReportExportData,
  TeacherClassReviewData,
  SkillProfileData,
  TeacherClassItem,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherInstanceItem,
  TeacherStudentItem,
  TimelineEvent,
} from './contracts'
import { normalizeSkillProfile, type RawSkillProfileResponse } from '@/utils/skillProfile'

interface RawTimelineItem {
  type: string
  challenge_id: string | number
  title: string
  detail?: string
  timestamp: string
  is_correct?: boolean
  points?: number
}

interface RawTimelineResponse {
  events: RawTimelineItem[]
}

interface RawTeacherEvidenceResponse {
  summary: {
    total_events: number
    proxy_request_count: number
    submit_count: number
    success_count: number
    challenge_id: string | number
  }
  events: Array<{
    type: string
    challenge_id: string | number
    title: string
    detail: string
    timestamp: string
    meta?: Record<string, unknown>
  }>
}

export async function getClasses(): Promise<TeacherClassItem[]> {
  return request<TeacherClassItem[]>({ method: 'GET', url: '/teacher/classes' })
}

export async function getClassStudents(
  name: string,
  params?: { keyword?: string; student_no?: string }
) {
  const payload = await request<
    Array<{
      id: string | number
      username: string
      student_no?: string
      name?: string
      solved_count?: number
      total_score?: number
      recent_event_count?: number
      weak_dimension?: string
    }>
  >({
    method: 'GET',
    url: `/teacher/classes/${encodeURIComponent(name)}/students`,
    params: {
      keyword: params?.keyword,
      student_no: params?.student_no,
    },
  })

  return payload.map((item) => ({
    ...item,
    id: String(item.id),
  }))
}

export async function getClassSummary(name: string): Promise<TeacherClassSummaryData> {
  return request<TeacherClassSummaryData>({
    method: 'GET',
    url: `/teacher/classes/${encodeURIComponent(name)}/summary`,
  })
}

export async function getClassTrend(name: string): Promise<TeacherClassTrendData> {
  return request<TeacherClassTrendData>({
    method: 'GET',
    url: `/teacher/classes/${encodeURIComponent(name)}/trend`,
  })
}

export async function getClassReview(name: string): Promise<TeacherClassReviewData> {
  const payload = await request<{
    class_name: string
    items: Array<{
      key: string
      title: string
      detail: string
      accent: 'danger' | 'warning' | 'success' | 'primary'
      students?: Array<{
        id: string | number
        username: string
        name?: string
      }>
      recommendation?: {
        challenge_id: string | number
        title: string
        category: RecommendationItem['category']
        difficulty: RecommendationItem['difficulty']
        reason: string
      }
    }>
  }>({
    method: 'GET',
    url: `/teacher/classes/${encodeURIComponent(name)}/review`,
  })

  return {
    ...payload,
    items: payload.items.map((item) => ({
      ...item,
      students: item.students?.map((student) => ({
        ...student,
        id: String(student.id),
      })),
      recommendation: item.recommendation
        ? {
            ...item.recommendation,
            challenge_id: String(item.recommendation.challenge_id),
          }
        : undefined,
    })),
  }
}

export async function getStudentProgress(id: string) {
  return request<MyProgressData>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/progress`,
  })
}

export async function getStudentSkillProfile(id: string): Promise<SkillProfileData> {
  const payload = await request<RawSkillProfileResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/skill-profile`,
  })
  return normalizeSkillProfile(payload)
}

export async function getStudentRecommendations(id: string): Promise<RecommendationItem[]> {
  const payload = await request<
    Array<{
      challenge_id: string | number
      title: string
      category: RecommendationItem['category']
      difficulty: RecommendationItem['difficulty']
      reason: string
    }>
  >({ method: 'GET', url: `/teacher/students/${encodeURIComponent(id)}/recommendations` })

  return payload.map((item) => ({
    ...item,
    challenge_id: String(item.challenge_id),
  }))
}

export async function getStudentTimeline(id: string): Promise<TimelineEvent[]> {
  const payload = await request<RawTimelineResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/timeline`,
  })

  return payload.events.map((item) => ({
    id: `${item.type}-${item.challenge_id}-${item.timestamp}`,
    type:
      item.type === 'instance_start' || item.type === 'instance_destroy'
        ? 'instance'
        : item.type === 'hint_unlock'
          ? 'hint'
        : item.type === 'flag_submit' && item.is_correct
          ? 'solve'
          : item.type === 'flag_submit'
            ? 'submit'
            : item.type,
    title: item.title,
    detail: item.detail,
    created_at: item.timestamp,
    challenge_id: String(item.challenge_id),
    is_correct: item.is_correct,
    points: item.points,
    meta: {
      raw_type: item.type,
    },
  }))
}

export async function getStudentEvidence(id: string): Promise<TeacherEvidenceData> {
  const payload = await request<RawTeacherEvidenceResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/evidence`,
  })

  return {
    summary: {
      ...payload.summary,
      challenge_id: String(payload.summary.challenge_id),
    },
    events: payload.events.map((item) => ({
      ...item,
      challenge_id: String(item.challenge_id),
    })),
  }
}

export async function getTeacherInstances(params?: {
  class_name?: string
  keyword?: string
  student_no?: string
}): Promise<TeacherInstanceItem[]> {
  const payload = await request<
    Array<{
      id: string | number
      student_id: string | number
      student_name: string
      student_username: string
      student_no?: string
      class_name: string
      challenge_id: string | number
      challenge_title: string
      status: string
      access_url?: string
      expires_at: string
      remaining_time: number
      extend_count: number
      max_extends: number
      created_at: string
    }>
  >({
    method: 'GET',
    url: '/teacher/instances',
    params: {
      class_name: params?.class_name,
      keyword: params?.keyword,
      student_no: params?.student_no,
    },
  })

  return payload.map((item) => ({
    ...item,
    id: String(item.id),
    student_id: String(item.student_id),
    challenge_id: String(item.challenge_id),
  }))
}

export async function destroyTeacherInstance(id: string): Promise<void> {
  return request<void>({ method: 'DELETE', url: `/teacher/instances/${encodeURIComponent(id)}` })
}

export async function exportClassReport(data: Record<string, unknown>) {
  return request<ReportExportData>({ method: 'POST', url: '/reports/class', data })
}
