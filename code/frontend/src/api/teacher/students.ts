import { request } from '../request'

import type {
  MyProgressData,
  RecommendationItem,
  ReportExportData,
  ReviewArchiveData,
  SkillProfileData,
  TeacherEvidenceData,
  TimelineEvent,
} from '../contracts'
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

interface RawReviewArchiveResponse {
  generated_at: string
  student: {
    id: string | number
    username: string
    name?: string
    class_name?: string
  }
  summary: ReviewArchiveData['summary']
  skill_profile?: RawSkillProfileResponse['dimensions']
  timeline: RawTimelineItem[]
  evidence: Array<{
    type: string
    challenge_id: string | number
    title: string
    detail?: string
    timestamp: string
    meta?: Record<string, unknown>
  }>
  writeups: Array<{
    id: string | number
    challenge_id: string | number
    challenge_title: string
    title: string
    submission_status: string
    visibility_status: string
    is_recommended: boolean
    published_at?: string
    updated_at: string
  }>
  manual_reviews: Array<{
    id: string | number
    challenge_id: string | number
    challenge_title: string
    answer: string
    review_status: string
    submitted_at: string
    reviewed_at?: string
    review_comment?: string
    score: number
    reviewer_name?: string
  }>
  teacher_observations: {
    items: Array<{
      key: string
      label: string
      level: string
      summary: string
      evidence?: string
    }>
  }
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

function normalizeTimelineEvent(item: RawTimelineItem): TimelineEvent {
  return {
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
  }
}

function normalizeReportExportData(
  payload: ReportExportData & { report_id: string | number }
): ReportExportData {
  return {
    ...payload,
    report_id: String(payload.report_id),
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

  return payload.events.map(normalizeTimelineEvent)
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

export async function exportStudentReviewArchive(
  studentId: string,
  data?: { format?: 'json' }
): Promise<ReportExportData> {
  const payload = await request<ReportExportData & { report_id: string | number }>({
    method: 'POST',
    url: `/teacher/students/${encodeURIComponent(studentId)}/review-archive/export`,
    data,
  })

  return normalizeReportExportData(payload)
}

export async function getStudentReviewArchive(studentId: string): Promise<ReviewArchiveData> {
  const payload = await request<RawReviewArchiveResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(studentId)}/review-archive`,
  })

  return {
    generated_at: payload.generated_at,
    student: {
      ...payload.student,
      id: String(payload.student.id),
    },
    summary: payload.summary,
    skill_profile: normalizeSkillProfile({
      user_id: payload.student.id,
      dimensions: payload.skill_profile ?? [],
      updated_at: payload.generated_at,
    }),
    timeline: payload.timeline.map(normalizeTimelineEvent),
    evidence: payload.evidence.map((item) => ({
      ...item,
      challenge_id: String(item.challenge_id),
    })),
    writeups: payload.writeups.map((item) => ({
      ...item,
      id: String(item.id),
      challenge_id: String(item.challenge_id),
    })),
    manual_reviews: payload.manual_reviews.map((item) => ({
      ...item,
      id: String(item.id),
      challenge_id: String(item.challenge_id),
    })),
    teacher_observations: payload.teacher_observations,
  }
}
