import { request } from '../request'

import type {
  PageResult,
  RecommendationItem,
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '../contracts'

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
      code: string
      severity: TeacherClassReviewData['items'][number]['severity']
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
