import { computed, ref, type Ref } from 'vue'
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router'

import type { AdminContestAWDServiceData, AWDCheckerType } from '@/api/contracts'

interface UseAwdChallengeSelectionOptions {
  contestId: Readonly<Ref<string>>
  route: RouteLocationNormalizedLoaded
  router: Router
  services: Ref<AdminContestAWDServiceData[]>
}

export function useAwdChallengeSelection(options: UseAwdChallengeSelectionOptions) {
  const { contestId, route, router, services } = options
  const selectedServiceId = ref('')

  const selectedService = computed(
    () => services.value.find((service) => service.id === selectedServiceId.value) || null
  )
  const selectedCheckerType = computed<AWDCheckerType | undefined>(
    () => selectedService.value?.checker_type
  )
  const sortedServices = computed(() =>
    [...services.value].sort(
      (left, right) =>
        left.order - right.order || left.display_name.localeCompare(right.display_name)
    )
  )

  function readServiceQuery(): string {
    const value = route.query.service
    if (Array.isArray(value)) {
      return String(value[0] ?? '')
    }
    return typeof value === 'string' ? value : ''
  }

  function syncServiceQuery(serviceId: string) {
    if (!serviceId || readServiceQuery() === serviceId) return
    void router.replace({
      name: 'ContestAWDConfig',
      params: { id: contestId.value },
      query: { ...route.query, service: serviceId },
    })
  }

  function reconcileSelectedServiceId() {
    const requestedServiceId = readServiceQuery()
    const selectedServiceStillExists = services.value.some(
      (service) => service.id === selectedServiceId.value
    )
    const requestedServiceExists = services.value.some(
      (service) => service.id === requestedServiceId
    )
    if (!selectedServiceId.value || !selectedServiceStillExists) {
      selectedServiceId.value = requestedServiceExists
        ? requestedServiceId
        : services.value[0]?.id || ''
    }
    syncServiceQuery(selectedServiceId.value)
  }

  function selectService(service: AdminContestAWDServiceData) {
    selectedServiceId.value = service.id
    void router.replace({
      name: 'ContestAWDConfig',
      params: { id: contestId.value },
      query: { service: service.id },
    })
  }

  return {
    selectedServiceId,
    selectedService,
    selectedCheckerType,
    sortedServices,
    readServiceQuery,
    syncServiceQuery,
    reconcileSelectedServiceId,
    selectService,
  }
}
