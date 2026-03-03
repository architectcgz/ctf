import { computed, onUnmounted, ref, watchEffect } from 'vue'

import { formatDurationHms } from '@/utils/format'

export function useCountdown(expiresAtIso: string | undefined) {
  const remainingSeconds = ref(0)

  let timer: number | undefined
  onUnmounted(() => {
    if (timer) window.clearInterval(timer)
  })

  watchEffect(() => {
    if (timer) window.clearInterval(timer)
    timer = undefined

    if (!expiresAtIso) {
      remainingSeconds.value = 0
      return
    }

    const expiresAt = new Date(expiresAtIso).getTime()
    if (Number.isNaN(expiresAt)) {
      remainingSeconds.value = 0
      return
    }

    const tick = () => {
      remainingSeconds.value = Math.max(0, Math.floor((expiresAt - Date.now()) / 1000))
    }
    tick()
    timer = window.setInterval(tick, 1000)
  })

  const formatted = computed(() => formatDurationHms(remainingSeconds.value))
  const isExpired = computed(() => remainingSeconds.value <= 0)
  const isUrgent = computed(() => remainingSeconds.value > 0 && remainingSeconds.value < 5 * 60)

  return { remainingSeconds, formatted, isExpired, isUrgent }
}

