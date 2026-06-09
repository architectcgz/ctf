import { computed, onMounted, onUnmounted, watch, type Ref } from 'vue'

const IMAGE_BUILD_POLL_INTERVAL_MS = 10_000

interface UseImageManageAutoRefreshOptions {
  hasActiveImages: Ref<boolean>
  refresh: () => Promise<void>
}

export function useImageManageAutoRefresh(options: UseImageManageAutoRefreshOptions) {
  const { hasActiveImages, refresh } = options

  let pollTimer: number | null = null

  function stopPolling() {
    if (pollTimer !== null) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  function startPolling() {
    if (pollTimer !== null) return
    pollTimer = window.setInterval(() => {
      void refresh()
    }, IMAGE_BUILD_POLL_INTERVAL_MS)
  }

  watch(
    hasActiveImages,
    (active) => {
      if (active) {
        startPolling()
        return
      }
      stopPolling()
    },
    { immediate: true }
  )

  onMounted(() => {
    void refresh()
  })

  onUnmounted(() => {
    stopPolling()
  })

  return {
    refreshHint: computed(() =>
      hasActiveImages.value ? '构建中镜像会每 10 秒自动刷新' : '当前无进行中镜像，可手动刷新'
    ),
  }
}
