import { request } from '../request'

import type {
  AdviceSeverity,
  MyProgressData,
  RecommendationData,
  ReportExportData,
  ReviewArchiveData,
  SkillProfileData,
  TeacherAttackSessionResponseData,
  TeacherEvidenceData,
  TimelineEvent,
} from '../contracts'
import { normalizeSkillProfile, type RawSkillProfileResponse } from '@/utils/skillProfile'
import { normalizeRecommendationData, type RawRecommendationResponse } from '@/utils/skillProfile'

export interface TeacherAttackSessionQuery {
  mode?: 'practice' | 'jeopardy' | 'awd'
  challenge_id?: string
  contest_id?: string
  round_id?: string
  result?: 'success' | 'failed' | 'in_progress' | 'unknown'
  with_events?: boolean
  limit?: number
  offset?: number
}

export interface TeacherEvidenceQuery {
  challenge_id?: string
  contest_id?: string
  round_id?: string
  event_type?: string
  from?: string
  to?: string
  limit?: number
  offset?: number
}

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
      code: string
      label?: string
      severity: AdviceSeverity
      dimension?: string
      summary: string
      evidence?: string
      action?: string
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

interface RawTeacherAttackSessionResponse extends Omit<
  TeacherAttackSessionResponseData,
  'sessions'
> {
  sessions: Array<
    Omit<
      TeacherAttackSessionResponseData['sessions'][number],
      | 'student_id'
      | 'team_id'
      | 'challenge_id'
      | 'contest_id'
      | 'round_id'
      | 'service_id'
      | 'victim_team_id'
      | 'events'
    > & {
      student_id: string | number
      team_id?: string | number
      challenge_id?: string | number
      contest_id?: string | number
      round_id?: string | number
      service_id?: string | number
      victim_team_id?: string | number
      events?: Array<
        Omit<
          NonNullable<TeacherAttackSessionResponseData['sessions'][number]['events']>[number],
          'actor' | 'target'
        > & {
          actor: {
            user_id: string | number
            team_id?: string | number
          }
          target: {
            challenge_id?: string | number
            contest_id?: string | number
            round_id?: string | number
            service_id?: string | number
            victim_team_id?: string | number
          }
        }
      >
    }
  >
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

export async function getStudentRecommendations(id: string): Promise<RecommendationData> {
  const payload = await request<RawRecommendationResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/recommendations`,
  })

  return normalizeRecommendationData(payload)
}

export async function getStudentTimeline(id: string): Promise<TimelineEvent[]> {
  const payload = await request<RawTimelineResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/timeline`,
  })

  return payload.events.map(normalizeTimelineEvent)
}

export async function getStudentEvidence(
  id: string,
  query: TeacherEvidenceQuery = {}
): Promise<TeacherEvidenceData> {
  const payload = await request<RawTeacherEvidenceResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/evidence`,
    params: {
      challenge_id: query.challenge_id,
      contest_id: query.contest_id,
      round_id: query.round_id,
      event_type: query.event_type,
      from: query.from,
      to: query.to,
      limit: query.limit,
      offset: query.offset,
    },
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

function optionalId(value: string | number | undefined): string | undefined {
  return value === undefined ? undefined : String(value)
}

export async function getStudentAttackSessions(
  id: string,
  query: TeacherAttackSessionQuery = {}
): Promise<TeacherAttackSessionResponseData> {
  const payload = await request<RawTeacherAttackSessionResponse>({
    method: 'GET',
    url: `/teacher/students/${encodeURIComponent(id)}/attack-sessions`,
    params: {
      mode: query.mode,
      challenge_id: query.challenge_id,
      contest_id: query.contest_id,
      round_id: query.round_id,
      result: query.result,
      with_events: query.with_events,
      limit: query.limit,
      offset: query.offset,
    },
  })

  return {
    summary: payload.summary,
    sessions: payload.sessions.map((session) => ({
      ...session,
      id: String(session.id),
      student_id: String(session.student_id),
      team_id: optionalId(session.team_id),
      challenge_id: optionalId(session.challenge_id),
      contest_id: optionalId(session.contest_id),
      round_id: optionalId(session.round_id),
      service_id: optionalId(session.service_id),
      victim_team_id: optionalId(session.victim_team_id),
      events: session.events?.map((event) => ({
        ...event,
        id: String(event.id),
        session_id: optionalId(event.session_id),
        actor: {
          user_id: String(event.actor.user_id),
          team_id: optionalId(event.actor.team_id),
        },
        target: {
          challenge_id: optionalId(event.target.challenge_id),
          contest_id: optionalId(event.target.contest_id),
          round_id: optionalId(event.target.round_id),
          service_id: optionalId(event.target.service_id),
          victim_team_id: optionalId(event.target.victim_team_id),
        },
      })),
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
