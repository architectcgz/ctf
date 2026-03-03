import { request } from './request'

import type { NotificationItem, PageResult } from './contracts'

export type GetNotificationsData = PageResult<NotificationItem>

export async function getNotifications(params?: Record<string, unknown>): Promise<GetNotificationsData> {
  return request<GetNotificationsData>({ method: 'GET', url: '/notifications', params })
}

export async function markAsRead(id: string) {
  return request<void>({ method: 'PUT', url: `/notifications/${encodeURIComponent(id)}/read` })
}
