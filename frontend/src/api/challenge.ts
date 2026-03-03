import { request } from './request'

export async function getChallenges(params?: Record<string, unknown>) {
  return request<{ items: unknown[]; total: number }>({ method: 'GET', url: '/challenges', params })
}

export async function getChallengeDetail(id: string) {
  return request<unknown>({ method: 'GET', url: `/challenges/${encodeURIComponent(id)}` })
}

export async function submitFlag(id: string, flag: string) {
  return request<unknown>({ method: 'POST', url: `/challenges/${encodeURIComponent(id)}/submissions`, data: { flag } })
}

export async function unlockHint(id: string, level: number) {
  return request<unknown>({ method: 'POST', url: `/challenges/${encodeURIComponent(id)}/hints/${level}/unlock` })
}

export async function createInstance(id: string) {
  return request<unknown>({ method: 'POST', url: `/challenges/${encodeURIComponent(id)}/instances` })
}

