import { request } from './request'

export async function getMyInstances() {
  return request<unknown[]>({ method: 'GET', url: '/instances' })
}

export async function destroyInstance(id: string) {
  return request<void>({ method: 'DELETE', url: `/instances/${encodeURIComponent(id)}` })
}

export async function extendInstance(id: string) {
  return request<void>({ method: 'POST', url: `/instances/${encodeURIComponent(id)}/extend` })
}

