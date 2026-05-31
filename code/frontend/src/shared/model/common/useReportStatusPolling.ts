import { onUnmounted, ref } from 'vue'

const POLL_INTERVAL_MS = 3000

export interface PollingReportStatus {
  report_id: string
  status: string
}

export function useReportStatusPolling<T extends PollingReportStatus>(
  fetchStatus: (reportId: string) => Promise<T>
) {
  const polling = ref(false)
  let timer: number | null = null

  async function pollOnce(
    reportId: string,
    onUpdate: (report: T) => void,
    onError?: (error: unknown) => void
  ) {
    try {
      const report = await fetchStatus(reportId)
      onUpdate(report)
      if (report.status !== 'processing') {
        stop()
      }
    } catch (error) {
      stop()
      console.error('轮询报告状态失败:', error)
      onError?.(error)
    }
  }

  function start(
    reportId: string,
    onUpdate: (report: T) => void,
    onError?: (error: unknown) => void
  ) {
    stop()
    polling.value = true

    timer = window.setInterval(() => {
      void pollOnce(reportId, onUpdate, onError)
    }, POLL_INTERVAL_MS)

    void pollOnce(reportId, onUpdate, onError)
  }

  function stop() {
    polling.value = false
    if (timer !== null) {
      window.clearInterval(timer)
      timer = null
    }
  }

  onUnmounted(() => {
    stop()
  })

  return {
    polling,
    start,
    stop,
  }
}
