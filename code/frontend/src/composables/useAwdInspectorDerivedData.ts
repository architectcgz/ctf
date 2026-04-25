import { computed, type Ref } from 'vue'

import type {
  AWDAttackLogData,
  AWDTrafficEventData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
} from '@/api/contracts'

interface AWDServiceAlertView {
  key: string
  label: string
  count: number
  affected_teams: string[]
  samples: Array<{
    service_id: string
    team_name: string
    challenge_title: string
  }>
}

interface UseAwdInspectorDerivedDataOptions {
  services: Ref<AWDTeamServiceData[]>
  attacks: Ref<AWDAttackLogData[]>
  trafficSummary: Ref<AWDTrafficSummaryData | null>
  trafficEvents: Ref<AWDTrafficEventData[]>
  serviceTeamFilter: Ref<string>
  serviceStatusFilter: Ref<'all' | AWDTeamServiceData['service_status']>
  serviceCheckSourceFilter: Ref<string>
  serviceAlertReasonFilter: Ref<string>
  attackTeamFilter: Ref<string>
  attackResultFilter: Ref<'all' | 'success' | 'failed'>
  attackSourceFilter: Ref<'all' | AWDAttackLogData['source']>
  getChallengeTitle: (challengeId: string) => string
  getCheckStatusLabel: (value: unknown) => string
}

export function useAwdInspectorDerivedData({
  services,
  attacks,
  trafficSummary,
  trafficEvents,
  serviceTeamFilter,
  serviceStatusFilter,
  serviceCheckSourceFilter,
  serviceAlertReasonFilter,
  attackTeamFilter,
  attackResultFilter,
  attackSourceFilter,
  getChallengeTitle,
  getCheckStatusLabel,
}: UseAwdInspectorDerivedDataOptions) {
  function getServiceCheckSourceValue(result: Record<string, unknown>): string {
    return typeof result.check_source === 'string' ? result.check_source : ''
  }

  function getServiceAlertReason(service: AWDTeamServiceData): string {
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

  const serviceTeamOptions = computed(() => {
    const seen = new Set<string>()
    return services.value.filter((item) => {
      if (seen.has(item.team_id)) {
        return false
      }
      seen.add(item.team_id)
      return true
    })
  })

  const serviceCheckSourceOptions = computed(() => {
    const seen = new Set<string>()
    return services.value
      .map((item) => getServiceCheckSourceValue(item.check_result))
      .filter((item) => {
        if (!item || seen.has(item)) {
          return false
        }
        seen.add(item)
        return true
      })
  })

  const baseFilteredServices = computed(() =>
    services.value.filter((item) => {
      if (serviceTeamFilter.value && item.team_id !== serviceTeamFilter.value) {
        return false
      }
      if (
        serviceStatusFilter.value !== 'all' &&
        item.service_status !== serviceStatusFilter.value
      ) {
        return false
      }
      if (
        serviceCheckSourceFilter.value &&
        getServiceCheckSourceValue(item.check_result) !== serviceCheckSourceFilter.value
      ) {
        return false
      }
      return true
    })
  )

  const serviceAlerts = computed<AWDServiceAlertView[]>(() => {
    const grouped = new Map<string, AWDServiceAlertView>()
    for (const service of baseFilteredServices.value) {
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
          challenge_title: getChallengeTitle(service.challenge_id),
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
  })

  const filteredServices = computed(() =>
    baseFilteredServices.value.filter((item) => {
      if (!serviceAlertReasonFilter.value) {
        return true
      }
      return getServiceAlertReason(item) === serviceAlertReasonFilter.value
    })
  )

  const attackTeamOptions = computed(() => {
    const seen = new Set<string>()
    return attacks.value.flatMap((item) => {
      const entries = [
        { id: item.attacker_team_id, name: item.attacker_team },
        { id: item.victim_team_id, name: item.victim_team },
      ]
      return entries.filter((entry) => {
        if (seen.has(entry.id)) {
          return false
        }
        seen.add(entry.id)
        return true
      })
    })
  })

  const attackSourceOptions = computed(() => {
    const seen = new Set<AWDAttackLogData['source']>()
    return attacks.value
      .map((item) => item.source)
      .filter((item) => {
        if (seen.has(item)) {
          return false
        }
        seen.add(item)
        return true
      })
  })

  const trafficTeamOptions = computed(() => {
    const entries = new Map<string, string>()

    for (const service of services.value) {
      entries.set(service.team_id, service.team_name)
    }
    for (const attackTeam of attackTeamOptions.value) {
      entries.set(attackTeam.id, attackTeam.name)
    }
    for (const item of trafficSummary.value?.top_attackers || []) {
      entries.set(item.team_id, item.team_name)
    }
    for (const item of trafficSummary.value?.top_victims || []) {
      entries.set(item.team_id, item.team_name)
    }
    for (const item of trafficEvents.value) {
      if (item.attacker_team_name?.trim()) {
        entries.set(item.attacker_team_id, item.attacker_team_name)
      }
      if (item.victim_team_name?.trim()) {
        entries.set(item.victim_team_id, item.victim_team_name)
      }
    }

    return [...entries.entries()]
      .map(([id, name]) => ({ id, name }))
      .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))
  })

  const filteredAttacks = computed(() =>
    attacks.value.filter((item) => {
      if (
        attackTeamFilter.value &&
        item.attacker_team_id !== attackTeamFilter.value &&
        item.victim_team_id !== attackTeamFilter.value
      ) {
        return false
      }
      if (attackResultFilter.value === 'success' && !item.is_success) {
        return false
      }
      if (attackResultFilter.value === 'failed' && item.is_success) {
        return false
      }
      if (attackSourceFilter.value !== 'all' && item.source !== attackSourceFilter.value) {
        return false
      }
      return true
    })
  )

  function getServiceAlertSubtitle(alert: AWDServiceAlertView): string {
    const teamLabel =
      alert.affected_teams.length === 0
        ? '无受影响队伍'
        : `影响队伍 ${alert.affected_teams.slice(0, 3).join(' / ')}`
    return `${teamLabel}${alert.affected_teams.length > 3 ? ' 等' : ''}`
  }

  function getServiceAlertClass(alertKey: string): string {
    switch (alertKey) {
      case 'invalid_access_url':
      case 'service_compromised':
        return 'awd-service-alert--danger'
      case 'unexpected_http_status':
      case 'http_request_failed':
      case 'all_probes_failed':
        return 'awd-service-alert--warning'
      default:
        return 'awd-service-alert--neutral'
    }
  }

  function getServiceAlertLabel(alertKey: string): string {
    return getCheckStatusLabel(alertKey) || alertKey
  }

  function applyServiceAlertFilter(alertKey: string): void {
    serviceAlertReasonFilter.value = serviceAlertReasonFilter.value === alertKey ? '' : alertKey
  }

  return {
    getServiceCheckSourceValue,
    serviceTeamOptions,
    serviceCheckSourceOptions,
    serviceAlerts,
    filteredServices,
    attackTeamOptions,
    attackSourceOptions,
    trafficTeamOptions,
    filteredAttacks,
    getServiceAlertReason,
    getServiceAlertSubtitle,
    getServiceAlertClass,
    getServiceAlertLabel,
    applyServiceAlertFilter,
  }
}
