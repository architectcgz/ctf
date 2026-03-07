import { onUnmounted, ref } from 'vue'

import { getReportStatus } from '@/api/assessment'
import type { ReportExportData } from '@/api/contracts'

const POLL_INTERVAL_MS = 3000

export function useReportStatusPolling() {
  const polling = ref(false)
  let timer: number | null = null

  async function pollOnce(reportId: string, onUpdate: (report: ReportExportData) => void) {
    const report = await getReportStatus(reportId)
    onUpdate(report)
    if (report.status !== 'processing') {
      stop()
    }
  }

  function start(reportId: string, onUpdate: (report: ReportExportData) => void) {
    stop()
    polling.value = true

    void pollOnce(reportId, onUpdate)

    timer = window.setInterval(() => {
      void pollOnce(reportId, onUpdate)
    }, POLL_INTERVAL_MS)
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
