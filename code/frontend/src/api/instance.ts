import { request } from './request'

import type { InstanceExtendData, InstanceListItem } from './contracts'

export async function getMyInstances(): Promise<InstanceListItem[]> {
  return request<InstanceListItem[]>({ method: 'GET', url: '/instances' })
}

export async function destroyInstance(id: string) {
  return request<void>({ method: 'DELETE', url: `/instances/${encodeURIComponent(id)}` })
}

export async function extendInstance(id: string) {
  return request<InstanceExtendData>({ method: 'POST', url: `/instances/${encodeURIComponent(id)}/extend` })
}
