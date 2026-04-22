import { onUnmounted, ref } from 'vue'

import { getReportStatus } from '@/api/assessment'
import type { ReportExportData } from '@/api/contracts'

const POLL_INTERVAL_MS = 3000

export function useReportStatusPolling() {
  const polling = ref(false)
  let timer: number | null = null

  async function pollOnce(
    reportId: string,
    onUpdate: (report: ReportExportData) => void,
    onError?: (error: unknown) => void
  ) {
    try {
      const report = await getReportStatus(reportId)
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
    onUpdate: (report: ReportExportData) => void,
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
