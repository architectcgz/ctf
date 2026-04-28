import { request } from './request'

import type {
  InstanceAccessInfo,
  InstanceData,
  InstanceExtendData,
  InstanceListItem,
} from './contracts'

interface RawExtendFields {
  remaining_extends?: number
  extend_count?: number
  max_extends?: number
}

export interface RawInstanceData
  extends Omit<InstanceData, 'id' | 'challenge_id' | 'remaining_extends'>, RawExtendFields {
  id: string | number
  challenge_id: string | number
}

interface RawInstanceListItem
  extends Omit<InstanceListItem, 'id' | 'challenge_id' | 'remaining_extends'>, RawExtendFields {
  id: string | number
  challenge_id: string | number
}

interface RawInstanceExtendData
  extends Omit<InstanceExtendData, 'id' | 'remaining_extends'>, RawExtendFields {
  id: string | number
}

function resolveRemainingExtends(item: RawExtendFields): number {
  if (typeof item.remaining_extends === 'number') {
    return Math.max(0, item.remaining_extends)
  }

  if (typeof item.max_extends === 'number' && typeof item.extend_count === 'number') {
    return Math.max(0, item.max_extends - item.extend_count)
  }

  return 0
}

function normalizeInstanceData(item: RawInstanceData): InstanceData {
  const { id, challenge_id, remaining_extends, extend_count, max_extends, ...rest } = item
  return {
    ...rest,
    id: String(id),
    challenge_id: String(challenge_id),
    share_scope: item.share_scope ?? 'per_user',
    remaining_extends: resolveRemainingExtends({ remaining_extends, extend_count, max_extends }),
  }
}

function normalizeInstanceListItem(item: RawInstanceListItem): InstanceListItem {
  const { challenge_title, category, difficulty } = item
  return {
    ...normalizeInstanceData(item),
    challenge_title,
    category,
    difficulty,
  }
}

export async function getMyInstances(): Promise<InstanceListItem[]> {
  const payload = await request<RawInstanceListItem[]>({ method: 'GET', url: '/instances' })
  return payload.map(normalizeInstanceListItem)
}

export async function destroyInstance(id: string) {
  return request<void>({
    method: 'DELETE',
    url: `/instances/${encodeURIComponent(id)}`,
    suppressErrorToast: true,
  })
}

export async function extendInstance(id: string) {
  const payload = await request<RawInstanceExtendData | null>({
    method: 'POST',
    url: `/instances/${encodeURIComponent(id)}/extend`,
  })

  if (!payload) {
    return null
  }

  return {
    id: String(payload.id),
    expires_at: payload.expires_at,
    remaining_extends: resolveRemainingExtends(payload),
  }
}

export { normalizeInstanceData }

export async function requestInstanceAccess(id: string) {
  return request<{ access_url: string; access?: InstanceAccessInfo }>({
    method: 'POST',
    url: `/instances/${encodeURIComponent(id)}/access`,
  })
}
