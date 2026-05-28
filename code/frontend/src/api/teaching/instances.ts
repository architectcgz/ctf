import { request } from '../request'

import type {
  ReportExportData,
  InstanceDirectorySummaryData,
  InstanceDirectoryPageData,
  ClassReportExportPayload,
  InstanceDirectoryItem,
} from '../contracts'

export type InstanceDirectoryStatusFilter =
  'running' | 'creating' | 'expired' | 'failed' | 'inactive'

export async function getInstanceDirectory(
  params?: {
    class_name?: string
    keyword?: string
    student_no?: string
    status?: InstanceDirectoryStatusFilter
    page?: number
    page_size?: number
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<InstanceDirectoryPageData<InstanceDirectoryItem>> {
  const payload = await request<
    InstanceDirectoryPageData<{
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
    summary: payload.summary as InstanceDirectorySummaryData,
    list: payload.list.map((item) => ({
      ...item,
      id: String(item.id),
      student_id: String(item.student_id),
      challenge_id: String(item.challenge_id),
    })),
  }
}

export async function destroyManagedInstance(id: string): Promise<void> {
  return request<void>({
    method: 'DELETE',
    url: `/teacher/instances/${encodeURIComponent(id)}`,
  })
}

export async function exportClassReport(data: ClassReportExportPayload) {
  return request<ReportExportData>({ method: 'POST', url: '/reports/class', data })
}
