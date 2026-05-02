import { ref, type Ref } from 'vue'

import type { ContestDetailData } from '@/api/contracts'
import { formatDuration } from '@/utils/format'

interface UseContestDetailCountdownOptions {
  contest: Ref<ContestDetailData | null>
}

export function useContestDetailCountdown({ contest }: UseContestDetailCountdownOptions) {
  const countdown = ref('')
  let countdownTimer: number | null = null

  function stopCountdown() {
    if (countdownTimer) {
      window.clearInterval(countdownTimer)
      countdownTimer = null
    }
  }

  function updateCountdown() {
    if (!contest.value) {
      countdown.value = ''
      stopCountdown()
      return
    }

    const now = Date.now()
    const start = new Date(contest.value.starts_at).getTime()
    const end = new Date(contest.value.ends_at).getTime()

    if (now < start) {
      countdown.value = `距离开始: ${formatDuration(start - now)}`
      return
    }
    if (now < end) {
      countdown.value = `距离结束: ${formatDuration(end - now)}`
      return
    }

    countdown.value = ''
    stopCountdown()
  }

  function startCountdown() {
    stopCountdown()
    if (!contest.value) {
      countdown.value = ''
      return
    }

    updateCountdown()
    if (countdown.value) {
      countdownTimer = window.setInterval(updateCountdown, 1000)
    }
  }

  return {
    countdown,
    startCountdown,
    stopCountdown,
  }
}
