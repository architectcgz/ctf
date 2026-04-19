import { request } from './request'

import type {
  TeacherEvidenceData,
  MyProgressData,
  PageResult,
  RecommendationItem,
  ReviewArchiveData,
  ReportExportData,
  SubmissionWriteupData,
  TeacherAWDReviewArchiveData,
  TeacherAWDReviewAttackItemData,
  TeacherAWDReviewContestItemData,
  TeacherAWDReviewRoundItemData,
  TeacherAWDReviewSelectedRoundData,
  TeacherAWDReviewServiceItemData,
  TeacherAWDReviewTeamItemData,
  TeacherAWDReviewTrafficItemData,
  TeacherClassReviewData,
  SkillProfileData,
  TeacherClassItem,
  TeacherClassSummaryData,
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
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

interface RawTeacherSubmissionWriteupItem extends Omit<
  TeacherSubmissionWriteupItemData,
  'id' | 'user_id' | 'challenge_id'
> {
  id: string | number
  user_id: string | number
  challenge_id: string | number
}

interface RawTeacherManualReviewSubmissionItem extends Omit<
  TeacherManualReviewSubmissionItemData,
  'id' | 'user_id' | 'challenge_id'
> {
  id: string | number
  user_id: string | number
  challenge_id: string | number
}

interface RawTeacherManualReviewSubmissionDetail extends Omit<
  TeacherManualReviewSubmissionDetailData,
  'id' | 'user_id' | 'challenge_id'
> {
  id: string | number
  user_id: string | number
  challenge_id: string | number
}

interface RawTeacherAWDReviewContestItem extends Omit<TeacherAWDReviewContestItemData, 'id'> {
  id: string | number
}

interface RawTeacherAWDReviewRoundItem extends Omit<
  TeacherAWDReviewRoundItemData,
  'id' | 'contest_id'
> {
  id: string | number
  contest_id: string | number
}

interface RawTeacherAWDReviewTeamItem extends Omit<
  TeacherAWDReviewTeamItemData,
  'team_id' | 'captain_id'
> {
  team_id: string | number
  captain_id: string | number
}

interface RawTeacherAWDReviewServiceItem extends Omit<
  TeacherAWDReviewServiceItemData,
  'id' | 'round_id' | 'team_id' | 'service_id' | 'challenge_id'
> {
  id: string | number
  round_id: string | number
  team_id: string | number
  service_id?: string | number
  challenge_id: string | number
}

interface RawTeacherAWDReviewAttackItem extends Omit<
  TeacherAWDReviewAttackItemData,
  'id' | 'round_id' | 'attacker_team_id' | 'victim_team_id' | 'service_id' | 'challenge_id'
> {
  id: string | number
  round_id: string | number
  attacker_team_id: string | number
  victim_team_id: string | number
  service_id?: string | number
  challenge_id: string | number
}

interface RawTeacherAWDReviewTrafficItem extends Omit<
  TeacherAWDReviewTrafficItemData,
  'id' | 'contest_id' | 'round_id' | 'attacker_team_id' | 'victim_team_id' | 'challenge_id'
> {
  id: string | number
  contest_id: string | number
  round_id: string | number
  attacker_team_id: string | number
  victim_team_id: string | number
  challenge_id: string | number
}

interface RawTeacherAWDReviewSelectedRound extends Omit<
  TeacherAWDReviewSelectedRoundData,
  'round' | 'teams' | 'services' | 'attacks' | 'traffic'
> {
  round: RawTeacherAWDReviewRoundItem
  teams: RawTeacherAWDReviewTeamItem[]
  services: RawTeacherAWDReviewServiceItem[]
  attacks: RawTeacherAWDReviewAttackItem[]
  traffic: RawTeacherAWDReviewTrafficItem[]
}

interface RawTeacherAWDReviewArchiveResponse extends Omit<
  TeacherAWDReviewArchiveData,
  'contest' | 'rounds' | 'selected_round' | 'scope'
> {
  scope: {
    snapshot_type: string
    requested_by: number
    requested_id: string | number
    requested_role?: string
  }
  contest: RawTeacherAWDReviewContestItem
  rounds: RawTeacherAWDReviewRoundItem[]
  selected_round?: RawTeacherAWDReviewSelectedRound
}

export async function getClasses(): Promise<TeacherClassItem[]>
export async function getClasses(params: {
  page?: number
  page_size?: number
}): Promise<PageResult<TeacherClassItem>>
export async function getClasses(params?: {
  page?: number
  page_size?: number
}): Promise<PageResult<TeacherClassItem> | TeacherClassItem[]> {
  const payload = await request<PageResult<TeacherClassItem>>({
    method: 'GET',
    url: '/teacher/classes',
    params: {
      page: params?.page,
      page_size: params?.page_size,
    },
  })

  return params ? payload : payload.list
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

export interface TeacherStudentDirectoryParams {
  class_name?: string
  keyword?: string
  student_no?: string
  sort_key?: 'name' | 'student_no' | 'total_score' | 'solved_count'
  sort_order?: 'asc' | 'desc'
  page?: number
  page_size?: number
  signal?: AbortSignal
}

export async function getStudentsDirectory(
  params?: TeacherStudentDirectoryParams
): Promise<PageResult<TeacherStudentItem>> {
  const payload = await request<
    PageResult<{
      id: string | number
      username: string
      student_no?: string
      name?: string
      class_name?: string
      solved_count?: number
      total_score?: number
      recent_event_count?: number
      weak_dimension?: string
    }>
  >({
    method: 'GET',
    url: '/teacher/students',
    params: {
      class_name: params?.class_name,
      keyword: params?.keyword,
      student_no: params?.student_no,
      sort_key: params?.sort_key,
      sort_order: params?.sort_order,
      page: params?.page,
      page_size: params?.page_size,
    },
    signal: params?.signal,
  })

  return {
    ...payload,
    list: payload.list.map((item) => ({
      ...item,
      id: String(item.id),
    })),
  }
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

export async function getTeacherWriteupSubmissions(params?: {
  student_id?: string
  challenge_id?: string
  class_name?: string
  submission_status?: 'draft' | 'published'
  visibility_status?: 'visible' | 'hidden'
  page?: number
  page_size?: number
}): Promise<PageResult<TeacherSubmissionWriteupItemData>> {
  const payload = await request<PageResult<RawTeacherSubmissionWriteupItem>>({
    method: 'GET',
    url: '/teacher/writeup-submissions',
    params: {
      student_id: params?.student_id,
      challenge_id: params?.challenge_id,
      class_name: params?.class_name,
      submission_status: params?.submission_status,
      visibility_status: params?.visibility_status,
      page: params?.page,
      page_size: params?.page_size,
    },
  })

  return {
    ...payload,
    list: payload.list.map((item) => ({
      ...item,
      id: String(item.id),
      user_id: String(item.user_id),
      challenge_id: String(item.challenge_id),
    })),
  }
}

function normalizeSubmissionWriteupData(
  item: SubmissionWriteupData & {
    id: string | number
    user_id: string | number
    challenge_id: string | number
    contest_id?: string | number
    recommended_by?: string | number
  }
): SubmissionWriteupData {
  return {
    ...item,
    id: String(item.id),
    user_id: String(item.user_id),
    challenge_id: String(item.challenge_id),
    contest_id: item.contest_id == null ? undefined : String(item.contest_id),
    recommended_by: item.recommended_by == null ? undefined : String(item.recommended_by),
  }
}

export async function recommendTeacherCommunityWriteup(id: string): Promise<SubmissionWriteupData> {
  const payload = await request<
    SubmissionWriteupData & {
      id: string | number
      user_id: string | number
      challenge_id: string | number
      contest_id?: string | number
      recommended_by?: string | number
    }
  >({
    method: 'POST',
    url: `/teacher/community-writeups/${encodeURIComponent(id)}/recommend`,
  })
  return normalizeSubmissionWriteupData(payload)
}

export async function unrecommendTeacherCommunityWriteup(
  id: string
): Promise<SubmissionWriteupData> {
  const payload = await request<
    SubmissionWriteupData & {
      id: string | number
      user_id: string | number
      challenge_id: string | number
      contest_id?: string | number
      recommended_by?: string | number
    }
  >({
    method: 'DELETE',
    url: `/teacher/community-writeups/${encodeURIComponent(id)}/recommend`,
  })
  return normalizeSubmissionWriteupData(payload)
}

export async function hideTeacherCommunityWriteup(id: string): Promise<SubmissionWriteupData> {
  const payload = await request<
    SubmissionWriteupData & {
      id: string | number
      user_id: string | number
      challenge_id: string | number
      contest_id?: string | number
      recommended_by?: string | number
    }
  >({
    method: 'POST',
    url: `/teacher/community-writeups/${encodeURIComponent(id)}/hide`,
  })
  return normalizeSubmissionWriteupData(payload)
}

export async function restoreTeacherCommunityWriteup(id: string): Promise<SubmissionWriteupData> {
  const payload = await request<
    SubmissionWriteupData & {
      id: string | number
      user_id: string | number
      challenge_id: string | number
      contest_id?: string | number
      recommended_by?: string | number
    }
  >({
    method: 'POST',
    url: `/teacher/community-writeups/${encodeURIComponent(id)}/restore`,
  })
  return normalizeSubmissionWriteupData(payload)
}

export async function getTeacherManualReviewSubmissions(params?: {
  student_id?: string
  challenge_id?: string
  class_name?: string
  review_status?: 'pending' | 'approved' | 'rejected'
  page?: number
  page_size?: number
}): Promise<PageResult<TeacherManualReviewSubmissionItemData>> {
  const payload = await request<PageResult<RawTeacherManualReviewSubmissionItem>>({
    method: 'GET',
    url: '/teacher/manual-review-submissions',
    params: {
      student_id: params?.student_id,
      challenge_id: params?.challenge_id,
      class_name: params?.class_name,
      review_status: params?.review_status,
      page: params?.page,
      page_size: params?.page_size,
    },
  })

  return {
    ...payload,
    list: payload.list.map((item) => ({
      ...item,
      id: String(item.id),
      user_id: String(item.user_id),
      challenge_id: String(item.challenge_id),
    })),
  }
}

export async function getTeacherManualReviewSubmission(
  id: string
): Promise<TeacherManualReviewSubmissionDetailData> {
  const payload = await request<RawTeacherManualReviewSubmissionDetail>({
    method: 'GET',
    url: `/teacher/manual-review-submissions/${encodeURIComponent(id)}`,
  })

  return {
    ...payload,
    id: String(payload.id),
    user_id: String(payload.user_id),
    challenge_id: String(payload.challenge_id),
  }
}

export async function reviewTeacherManualReviewSubmission(
  id: string,
  payload: {
    review_status: 'approved' | 'rejected'
    review_comment?: string
  }
): Promise<TeacherManualReviewSubmissionDetailData> {
  const response = await request<RawTeacherManualReviewSubmissionDetail>({
    method: 'PUT',
    url: `/teacher/manual-review-submissions/${encodeURIComponent(id)}/review`,
    data: payload,
  })

  return {
    ...response,
    id: String(response.id),
    user_id: String(response.user_id),
    challenge_id: String(response.challenge_id),
  }
}

export async function getTeacherInstances(params?: {
  class_name?: string
  keyword?: string
  student_no?: string
}, options?: {
  signal?: AbortSignal
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
    signal: options?.signal,
  })

  return payload.map((item) => ({
    ...item,
    id: String(item.id),
    student_id: String(item.student_id),
    challenge_id: String(item.challenge_id),
  }))
}

export async function destroyTeacherInstance(id: string): Promise<void> {
  return request<void>({
    method: 'DELETE',
    url: `/teacher/instances/${encodeURIComponent(id)}`,
    suppressErrorToast: true,
  })
}

export async function exportClassReport(data: Record<string, unknown>) {
  return request<ReportExportData>({ method: 'POST', url: '/reports/class', data })
}

export async function listTeacherAWDReviews(params?: {
  status?: TeacherAWDReviewContestItemData['status']
  keyword?: string
}): Promise<TeacherAWDReviewContestItemData[]> {
  const payload = await request<{ contests: RawTeacherAWDReviewContestItem[] }>({
    method: 'GET',
    url: '/teacher/awd/reviews',
    params: {
      status: params?.status,
      keyword: params?.keyword,
    },
  })

  return payload.contests.map(normalizeTeacherAWDReviewContest)
}

export async function getTeacherAWDReview(
  contestId: string,
  params?: {
    round?: number
    team_id?: string
  }
): Promise<TeacherAWDReviewArchiveData> {
  const payload = await request<RawTeacherAWDReviewArchiveResponse>({
    method: 'GET',
    url: `/teacher/awd/reviews/${encodeURIComponent(contestId)}`,
    params: {
      round: params?.round,
      team_id: params?.team_id,
    },
  })

  return {
    ...payload,
    scope: {
      ...payload.scope,
      requested_id: String(payload.scope.requested_id),
    },
    contest: normalizeTeacherAWDReviewContest(payload.contest),
    rounds: payload.rounds.map(normalizeTeacherAWDReviewRound),
    selected_round: payload.selected_round
      ? normalizeTeacherAWDSelectedRound(payload.selected_round)
      : undefined,
  }
}

export async function exportTeacherAWDReviewArchive(
  contestId: string,
  data?: { round_number?: number }
): Promise<ReportExportData> {
  const payload = await request<ReportExportData & { report_id: string | number }>({
    method: 'POST',
    url: `/teacher/awd/reviews/${encodeURIComponent(contestId)}/export/archive`,
    data,
  })

  return normalizeReportExportData(payload)
}

export async function exportTeacherAWDReviewReport(
  contestId: string,
  data?: { round_number?: number }
): Promise<ReportExportData> {
  const payload = await request<ReportExportData & { report_id: string | number }>({
    method: 'POST',
    url: `/teacher/awd/reviews/${encodeURIComponent(contestId)}/export/report`,
    data,
  })

  return normalizeReportExportData(payload)
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

function normalizeTeacherAWDReviewContest(
  item: RawTeacherAWDReviewContestItem
): TeacherAWDReviewContestItemData {
  return {
    ...item,
    id: String(item.id),
  }
}

function normalizeTeacherAWDReviewRound(
  item: RawTeacherAWDReviewRoundItem
): TeacherAWDReviewRoundItemData {
  return {
    ...item,
    id: String(item.id),
    contest_id: String(item.contest_id),
  }
}

function normalizeTeacherAWDReviewTeam(
  item: RawTeacherAWDReviewTeamItem
): TeacherAWDReviewTeamItemData {
  return {
    ...item,
    team_id: String(item.team_id),
    captain_id: String(item.captain_id),
  }
}

function normalizeTeacherAWDReviewService(
  item: RawTeacherAWDReviewServiceItem
): TeacherAWDReviewServiceItemData {
  return {
    ...item,
    id: String(item.id),
    round_id: String(item.round_id),
    team_id: String(item.team_id),
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeTeacherAWDReviewAttack(
  item: RawTeacherAWDReviewAttackItem
): TeacherAWDReviewAttackItemData {
  return {
    ...item,
    id: String(item.id),
    round_id: String(item.round_id),
    attacker_team_id: String(item.attacker_team_id),
    victim_team_id: String(item.victim_team_id),
    service_id: item.service_id == null ? undefined : String(item.service_id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeTeacherAWDReviewTraffic(
  item: RawTeacherAWDReviewTrafficItem
): TeacherAWDReviewTrafficItemData {
  return {
    ...item,
    id: String(item.id),
    contest_id: String(item.contest_id),
    round_id: String(item.round_id),
    attacker_team_id: String(item.attacker_team_id),
    victim_team_id: String(item.victim_team_id),
    challenge_id: String(item.challenge_id),
  }
}

function normalizeTeacherAWDSelectedRound(
  item: RawTeacherAWDReviewSelectedRound
): TeacherAWDReviewSelectedRoundData {
  return {
    round: normalizeTeacherAWDReviewRound(item.round),
    teams: item.teams.map(normalizeTeacherAWDReviewTeam),
    services: item.services.map(normalizeTeacherAWDReviewService),
    attacks: item.attacks.map(normalizeTeacherAWDReviewAttack),
    traffic: item.traffic.map(normalizeTeacherAWDReviewTraffic),
  }
}
