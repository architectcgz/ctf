import { computed, type Ref } from 'vue'

import type {
  AWDAttackLogData,
  AWDTrafficEventData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
} from '@/api/contracts'
import {
  buildAttackSourceOptions,
  buildAttackTeamOptions,
  buildTrafficTeamOptions,
  filterAttacks,
} from './awdInspectorAttackDerived'
import {
  buildServiceAlerts,
  buildServiceCheckSourceOptions,
  buildServiceTeamOptions,
  filterServices,
  getServiceAlertReason,
  getServiceCheckSourceValue,
  type AWDServiceAlertView,
} from './awdInspectorServiceDerived'

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
  const serviceTeamOptions = computed(() => buildServiceTeamOptions(services.value))

  const serviceCheckSourceOptions = computed(() => buildServiceCheckSourceOptions(services.value))

  const baseFilteredServices = computed(() =>
    filterServices(services.value, {
      serviceTeamFilter: serviceTeamFilter.value,
      serviceStatusFilter: serviceStatusFilter.value,
      serviceCheckSourceFilter: serviceCheckSourceFilter.value,
    })
  )

  const serviceAlerts = computed<AWDServiceAlertView[]>(() =>
    buildServiceAlerts(baseFilteredServices.value, getChallengeTitle, getCheckStatusLabel)
  )

  const filteredServices = computed(() =>
    baseFilteredServices.value.filter((item) => {
      if (!serviceAlertReasonFilter.value) {
        return true
      }
      return getServiceAlertReason(item) === serviceAlertReasonFilter.value
    })
  )

  const attackTeamOptions = computed(() => buildAttackTeamOptions(attacks.value))

  const attackSourceOptions = computed(() => buildAttackSourceOptions(attacks.value))

  const trafficTeamOptions = computed(() =>
    buildTrafficTeamOptions({
      services: services.value,
      attackTeamOptions: attackTeamOptions.value,
      trafficSummary: trafficSummary.value,
      trafficEvents: trafficEvents.value,
    })
  )

  const filteredAttacks = computed(() =>
    filterAttacks(attacks.value, {
      attackTeamFilter: attackTeamFilter.value,
      attackResultFilter: attackResultFilter.value,
      attackSourceFilter: attackSourceFilter.value,
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
