import { request } from './request'

export async function getSkillProfile() {
  return request<unknown>({ method: 'GET', url: '/users/me/skill-profile' })
}

export async function getRecommendations() {
  return request<unknown>({ method: 'GET', url: '/users/me/recommendations' })
}

export async function getMyProgress() {
  return request<unknown>({ method: 'GET', url: '/users/me/progress' })
}

export async function getMyTimeline() {
  return request<unknown>({ method: 'GET', url: '/users/me/timeline' })
}

export async function exportPersonalReport(data: Record<string, unknown>) {
  return request<unknown>({ method: 'POST', url: '/reports/personal', data })
}

