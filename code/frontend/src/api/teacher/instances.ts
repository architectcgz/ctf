import { request } from '../request'

import type {
  ReportExportData,
  TeacherInstanceListSummaryData,
  TeacherInstancePageData,
  TeacherClassReportExportPayload,
  TeacherInstanceItem,
} from '../contracts'

export type TeacherInstanceStatusFilter = 'running' | 'creating' | 'expired' | 'failed' | 'inactive'

export async function getTeacherInstances(
  params?: {
    class_name?: string
    keyword?: string
    student_no?: string
    status?: TeacherInstanceStatusFilter
    page?: number
    page_size?: number
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<TeacherInstancePageData<TeacherInstanceItem>> {
  const payload = await request<
    TeacherInstancePageData<{
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
      status: params?.status,
      page: params?.page,
      page_size: params?.page_size,
    },
    signal: options?.signal,
  })

  return {
    ...payload,
    summary: payload.summary as TeacherInstanceListSummaryData,
    list: payload.list.map((item) => ({
      ...item,
      id: String(item.id),
      student_id: String(item.student_id),
      challenge_id: String(item.challenge_id),
    })),
  }
}

export async function destroyTeacherInstance(id: string): Promise<void> {
  return request<void>({
    method: 'DELETE',
    url: `/teacher/instances/${encodeURIComponent(id)}`,
  })
}

export async function exportClassReport(data: TeacherClassReportExportPayload) {
  return request<ReportExportData>({ method: 'POST', url: '/reports/class', data })
}
