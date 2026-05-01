import { request } from '../request'

import type {
  AdminCheatDetectionData,
  AdminDashboardData,
  AdminNotificationPublishPayload,
  AdminNotificationPublishResult,
  AuditLogItem,
  PageResult,
} from '../contracts'

interface RawAdminNotificationPublishResult {
  batch_id: string | number
  recipient_count: number
}

interface RawCheatDetectionData {
  generated_at: string
  summary: {
    submit_burst_users: number
    shared_ip_groups: number
    affected_users: number
  }
  suspects: Array<{
    user_id: string | number
    username: string
    submit_count: number
    last_seen_at: string
    reason: string
  }>
  shared_ips: Array<{
    ip: string
    user_count: number
    usernames: string[]
  }>
}

function normalizeCheatDetection(data: RawCheatDetectionData): AdminCheatDetectionData {
  return {
    generated_at: data.generated_at,
    summary: data.summary,
    suspects: data.suspects.map((item) => ({
      ...item,
      user_id: String(item.user_id),
    })),
    shared_ips: data.shared_ips,
  }
}

export async function getDashboard(): Promise<AdminDashboardData> {
  return request<AdminDashboardData>({ method: 'GET', url: '/admin/dashboard' })
}

export async function publishAdminNotification(
  data: AdminNotificationPublishPayload
): Promise<AdminNotificationPublishResult> {
  const response = await request<RawAdminNotificationPublishResult>({
    method: 'POST',
    url: '/admin/notifications',
    data,
  })

  return {
    batch_id: String(response.batch_id),
    recipient_count: response.recipient_count,
  }
}

export async function getAuditLogs(
  params?: Record<string, unknown>,
  options?: { signal?: AbortSignal }
) {
  return request<PageResult<AuditLogItem>>({
    method: 'GET',
    url: '/admin/audit-logs',
    params,
    signal: options?.signal,
  })
}

export async function getCheatDetection(): Promise<AdminCheatDetectionData> {
  const response = await request<RawCheatDetectionData>({
    method: 'GET',
    url: '/admin/cheat-detection',
  })
  return normalizeCheatDetection(response)
}
