import type { AWDTeamServiceData } from '@/api/contracts'

export interface AWDServiceAlertView {
  key: string
  label: string
  count: number
  affected_teams: string[]
  samples: Array<{
    service_id: string
    team_name: string
    awd_challenge_title: string
  }>
}

export function getServiceCheckSourceValue(result: Record<string, unknown>): string {
  return typeof result.check_source === 'string' ? result.check_source : ''
}

export function getServiceAlertReason(service: AWDTeamServiceData): string {
  if (service.service_status === 'up') {
    return ''
  }
  const errorCode =
    typeof service.check_result.error_code === 'string'
      ? service.check_result.error_code.trim()
      : ''
  if (errorCode) {
    return errorCode
  }
  const statusReason =
    typeof service.check_result.status_reason === 'string'
      ? service.check_result.status_reason.trim()
      : ''
  if (statusReason && statusReason !== 'healthy') {
    return statusReason
  }
  if (service.service_status === 'compromised') {
    return 'service_compromised'
  }
  if (service.service_status === 'down') {
    return 'service_down'
  }
  return ''
}

export function buildServiceTeamOptions(services: AWDTeamServiceData[]): AWDTeamServiceData[] {
  const seen = new Set<string>()
  return services.filter((item) => {
    if (seen.has(item.team_id)) {
      return false
    }
    seen.add(item.team_id)
    return true
  })
}

export function buildServiceCheckSourceOptions(services: AWDTeamServiceData[]): string[] {
  const seen = new Set<string>()
  return services
    .map((item) => getServiceCheckSourceValue(item.check_result))
    .filter((item) => {
      if (!item || seen.has(item)) {
        return false
      }
      seen.add(item)
      return true
    })
}

interface ServiceFilterOptions {
  serviceTeamFilter: string
  serviceStatusFilter: 'all' | AWDTeamServiceData['service_status']
  serviceCheckSourceFilter: string
}

export function filterServices(
  services: AWDTeamServiceData[],
  { serviceTeamFilter, serviceStatusFilter, serviceCheckSourceFilter }: ServiceFilterOptions
): AWDTeamServiceData[] {
  return services.filter((item) => {
    if (serviceTeamFilter && item.team_id !== serviceTeamFilter) {
      return false
    }
    if (serviceStatusFilter !== 'all' && item.service_status !== serviceStatusFilter) {
      return false
    }
    if (
      serviceCheckSourceFilter &&
      getServiceCheckSourceValue(item.check_result) !== serviceCheckSourceFilter
    ) {
      return false
    }
    return true
  })
}

export function buildServiceAlerts(
  services: AWDTeamServiceData[],
  getChallengeTitle: (challengeId: string) => string,
  getCheckStatusLabel: (value: unknown) => string
): AWDServiceAlertView[] {
  const grouped = new Map<string, AWDServiceAlertView>()
  for (const service of services) {
    const reason = getServiceAlertReason(service)
    if (!reason) {
      continue
    }
    const existing = grouped.get(reason) || {
      key: reason,
      label: getCheckStatusLabel(reason) || reason,
      count: 0,
      affected_teams: [],
      samples: [],
    }
    existing.count += 1
    if (!existing.affected_teams.includes(service.team_name)) {
      existing.affected_teams.push(service.team_name)
    }
    if (existing.samples.length < 3) {
      existing.samples.push({
        service_id: service.service_id || '',
        team_name: service.team_name,
        awd_challenge_title: getChallengeTitle(service.awd_challenge_id),
      })
    }
    grouped.set(reason, existing)
  }

  return [...grouped.values()].sort((left, right) => {
    if (left.count !== right.count) {
      return right.count - left.count
    }
    return left.label.localeCompare(right.label, 'zh-CN')
  })
}
