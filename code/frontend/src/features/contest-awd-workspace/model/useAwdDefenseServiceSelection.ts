import { ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import type { AWDDefenseServiceCard } from './awdDefensePresentation'

export function useAwdDefenseServiceSelection(
  services: MaybeRefOrGetter<AWDDefenseServiceCard[]>
) {
  const selectedServiceId = ref('')

  function selectService(serviceId: string): void {
    selectedServiceId.value = serviceId
  }

  watch(
    () => toValue(services).map((service) => service.serviceId),
    (serviceIds) => {
      if (serviceIds.length === 0) {
        selectedServiceId.value = ''
        return
      }
      if (!serviceIds.includes(selectedServiceId.value)) {
        selectedServiceId.value = serviceIds[0]
      }
    },
    { immediate: true }
  )

  return {
    selectedServiceId,
    selectService,
  }
}
