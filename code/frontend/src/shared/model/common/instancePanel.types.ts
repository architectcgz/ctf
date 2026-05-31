export const instancePanelStatuses = [
  'pending',
  'creating',
  'running',
  'expired',
  'destroying',
  'destroyed',
  'failed',
  'crashed',
] as const

export const instancePanelShareScopes = ['shared', 'per_team', 'per_user'] as const

export type InstancePanelStatus = (typeof instancePanelStatuses)[number]
export type InstancePanelShareScope = (typeof instancePanelShareScopes)[number]

export interface InstancePanelItem {
  id: string
  challenge_title: string
  status: InstancePanelStatus
  access_url?: string | null
  share_scope: InstancePanelShareScope
  remaining_extends: number
  expires_at: string
}
