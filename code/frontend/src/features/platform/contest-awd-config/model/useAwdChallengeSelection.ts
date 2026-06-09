import { computed, ref, type Ref } from 'vue'

import type { AdminContestAWDServiceData, AWDCheckerType } from '@/api/contracts'

interface UseAwdChallengeSelectionOptions {
  services: Ref<AdminContestAWDServiceData[]>
  readServiceQuery: () => string
  replaceServiceQuery: (serviceId: string) => void
}

export function useAwdChallengeSelection(options: UseAwdChallengeSelectionOptions) {
  const { services, readServiceQuery, replaceServiceQuery } = options
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

  function syncServiceQuery(serviceId: string) {
    if (!serviceId || readServiceQuery() === serviceId) return
    replaceServiceQuery(serviceId)
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
    syncServiceQuery(service.id)
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
