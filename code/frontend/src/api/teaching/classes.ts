import { request } from '../request'

import type {
  ClassDirectoryItem,
  PageResult,
  RecommendationItem,
  ClassInsightQueryData,
  TeacherOverviewData,
  ClassInsightReviewData,
  ClassInsightSummaryData,
  ClassInsightTrendData,
  StudentDirectoryItem,
} from '../contracts'

export async function getClasses(): Promise<ClassDirectoryItem[]>
export async function getClasses(params: {
  page?: number
  page_size?: number
}): Promise<PageResult<ClassDirectoryItem>>
export async function getClasses(params?: {
  page?: number
  page_size?: number
}): Promise<PageResult<ClassDirectoryItem> | ClassDirectoryItem[]> {
  const payload = await request<PageResult<ClassDirectoryItem>>({
    method: 'GET',
    url: '/teacher/classes',
    params: {
      page: params?.page,
      page_size: params?.page_size,
    },
  })

  return params ? payload : payload.list
}

function normalizeStudentDirectoryItem(item: {
  id: string | number
  username: string
  student_no?: string
  name?: string
  class_name?: string
  solved_count?: number
  total_score?: number
  recent_event_count?: number
  weak_dimension?: string
}) {
  return {
    ...item,
    id: String(item.id),
  }
}

export async function getTeacherOverview(): Promise<TeacherOverviewData> {
  const payload = await request<{
    summary: TeacherOverviewData['summary']
    trend: TeacherOverviewData['trend']
    focus_classes: TeacherOverviewData['focus_classes']
    focus_students: Array<{
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
    spotlight_student?: {
      id: string | number
      username: string
      student_no?: string
      name?: string
      class_name?: string
      solved_count?: number
      total_score?: number
      recent_event_count?: number
      weak_dimension?: string
    } | null
    weak_dimensions: TeacherOverviewData['weak_dimensions']
  }>({
    method: 'GET',
    url: '/teacher/overview',
  })

  return {
    ...payload,
    focus_students: payload.focus_students.map(normalizeStudentDirectoryItem),
    spotlight_student: payload.spotlight_student
      ? normalizeStudentDirectoryItem(payload.spotlight_student)
      : null,
  }
}

export async function getClassStudents(
  name: string,
  params?: { keyword?: string; student_no?: string }
): Promise<StudentDirectoryItem[]> {
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

  return payload.map(normalizeStudentDirectoryItem)
}

export interface StudentDirectoryParams {
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
  params?: StudentDirectoryParams
): Promise<PageResult<StudentDirectoryItem>> {
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
    list: payload.list.map(normalizeStudentDirectoryItem),
  }
}

export async function getClassSummary(
  name: string,
  params?: ClassInsightQueryData
): Promise<ClassInsightSummaryData> {
  return request<ClassInsightSummaryData>({
    method: 'GET',
    url: `/teacher/classes/${encodeURIComponent(name)}/summary`,
    params,
  })
}

export async function getClassTrend(
  name: string,
  params?: ClassInsightQueryData
): Promise<ClassInsightTrendData> {
  return request<ClassInsightTrendData>({
    method: 'GET',
    url: `/teacher/classes/${encodeURIComponent(name)}/trend`,
    params,
  })
}

export async function getClassReview(
  name: string,
  params?: ClassInsightQueryData
): Promise<ClassInsightReviewData> {
  const payload = await request<{
    class_name: string
    items: Array<{
      code: string
      severity: ClassInsightReviewData['items'][number]['severity']
      summary: string
      evidence?: string
      action?: string
      reason_codes?: string[]
      dimension?: string
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
        dimension?: string
        difficulty_band?: RecommendationItem['difficulty_band']
        severity?: RecommendationItem['severity']
        reason_codes?: string[]
        summary: string
        evidence?: string
      }
    }>
  }>({
    method: 'GET',
    url: `/teacher/classes/${encodeURIComponent(name)}/review`,
    params,
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
