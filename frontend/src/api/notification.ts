import { request } from './request'

export async function getNotifications(params?: Record<string, unknown>) {
  return request<{ items: unknown[]; total: number }>({ method: 'GET', url: '/notifications', params })
}

export async function markAsRead(id: string) {
  return request<void>({ method: 'PUT', url: `/notifications/${encodeURIComponent(id)}/read` })
}

