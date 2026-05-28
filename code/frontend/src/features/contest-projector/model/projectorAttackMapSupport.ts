import type { AWDTeamServiceData } from '@/api/contracts'

export type AttackMapDetailPanel = 'teams' | 'ranking' | 'attacks'

export type ProjectorServiceIconName = 'database' | 'globe' | 'server' | 'shield'

export function getProjectorAttackServiceKey(teamId: string, service: AWDTeamServiceData): string {
  return service.service_id
    ? `${teamId}:service:${service.service_id}`
    : `${teamId}:challenge:${service.awd_challenge_id}`
}

export function getProjectorServiceDisplayName(service: AWDTeamServiceData): string {
  return (
    service.service_name?.trim() ||
    service.awd_challenge_title?.trim() ||
    (service.service_id ? `服务 ${service.service_id}` : `题目 ${service.awd_challenge_id}`)
  )
}

export function getProjectorServiceIconName(
  service: AWDTeamServiceData
): ProjectorServiceIconName {
  const label = getProjectorServiceDisplayName(service).toLowerCase()
  if (
    label.includes('drive') ||
    label.includes('盘') ||
    label.includes('data') ||
    label.includes('db')
  ) {
    return 'database'
  }
  if (label.includes('web') || label.includes('ticket') || label.includes('工单')) {
    return 'globe'
  }
  if (service.service_status === 'compromised' || service.service_status === 'down') {
    return 'shield'
  }
  return 'server'
}
