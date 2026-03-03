import { onUnmounted, ref } from 'vue'

export function useAbortController() {
  const controller = ref<AbortController | null>(null)

  const createController = () => {
    controller.value?.abort()
    controller.value = new AbortController()
    return controller.value
  }

  const abort = () => {
    controller.value?.abort()
    controller.value = null
  }

  onUnmounted(() => {
    abort()
  })

  return {
    createController,
    abort,
    signal: () => controller.value?.signal,
  }
}
