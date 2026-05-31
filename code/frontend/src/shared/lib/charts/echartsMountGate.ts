import { onBeforeUnmount, onMounted, ref } from 'vue'

export function useEChartsMountGate() {
  const containerRef = ref<HTMLElement | null>(null)
  const isChartReady = ref(false)
  let resizeObserver: ResizeObserver | null = null
  let frameId: number | null = null

  function clearFrame(): void {
    if (frameId === null || typeof window === 'undefined') {
      return
    }
    window.cancelAnimationFrame(frameId)
    frameId = null
  }

  function disconnectResizeObserver(): void {
    resizeObserver?.disconnect()
    resizeObserver = null
  }

  function updateReadyState(): void {
    if (isChartReady.value) {
      disconnectResizeObserver()
      return
    }

    const container = containerRef.value
    if (!container) {
      return
    }

    if (container.clientWidth > 0 && container.clientHeight > 0) {
      isChartReady.value = true
      disconnectResizeObserver()
    }
  }

  function scheduleReadyCheck(): void {
    if (typeof window === 'undefined') {
      return
    }

    clearFrame()
    frameId = window.requestAnimationFrame(() => {
      frameId = null
      updateReadyState()
    })
  }

  onMounted(() => {
    scheduleReadyCheck()

    if (typeof ResizeObserver === 'undefined') {
      return
    }

    resizeObserver = new ResizeObserver(() => {
      updateReadyState()
    })

    if (containerRef.value) {
      resizeObserver.observe(containerRef.value)
    }
  })

  onBeforeUnmount(() => {
    clearFrame()
    disconnectResizeObserver()
  })

  return {
    containerRef,
    isChartReady,
  }
}
