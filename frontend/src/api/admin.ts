import { request } from './request'

export async function getDashboard() {
  return request<unknown>({ method: 'GET', url: '/admin/dashboard' })
}

export async function getUsers(params?: Record<string, unknown>) {
  return request<{ items: unknown[]; total: number }>({ method: 'GET', url: '/admin/users', params })
}

export async function createUser(data: Record<string, unknown>) {
  return request<unknown>({ method: 'POST', url: '/admin/users', data })
}

export async function updateUser(id: string, data: Record<string, unknown>) {
  return request<unknown>({ method: 'PUT', url: `/admin/users/${encodeURIComponent(id)}`, data })
}

export async function deleteUser(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/users/${encodeURIComponent(id)}` })
}

export async function importUsers(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request<unknown>({
    method: 'POST',
    url: '/admin/users/import',
    data: form,
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function getChallenges(params?: Record<string, unknown>) {
  return request<{ items: unknown[]; total: number }>({ method: 'GET', url: '/admin/challenges', params })
}

export async function createChallenge(data: Record<string, unknown>) {
  return request<unknown>({ method: 'POST', url: '/admin/challenges', data })
}

export async function updateChallenge(id: string, data: Record<string, unknown>) {
  return request<unknown>({ method: 'PUT', url: `/admin/challenges/${encodeURIComponent(id)}`, data })
}

export async function deleteChallenge(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/challenges/${encodeURIComponent(id)}` })
}

export async function getImages(params?: Record<string, unknown>) {
  return request<{ items: unknown[]; total: number }>({ method: 'GET', url: '/admin/images', params })
}

export async function createImage(data: Record<string, unknown>) {
  return request<unknown>({ method: 'POST', url: '/admin/images', data })
}

export async function deleteImage(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/images/${encodeURIComponent(id)}` })
}

export async function getAuditLogs(params?: Record<string, unknown>) {
  return request<{ items: unknown[]; total: number }>({ method: 'GET', url: '/admin/audit-logs', params })
}

export async function getContests(params?: Record<string, unknown>) {
  return request<{ items: unknown[]; total: number }>({ method: 'GET', url: '/admin/contests', params })
}

export async function createContest(data: Record<string, unknown>) {
  return request<unknown>({ method: 'POST', url: '/admin/contests', data })
}

export async function updateContest(id: string, data: Record<string, unknown>) {
  return request<unknown>({ method: 'PUT', url: `/admin/contests/${encodeURIComponent(id)}`, data })
}

export async function deleteContest(id: string) {
  return request<void>({ method: 'DELETE', url: `/admin/contests/${encodeURIComponent(id)}` })
}

