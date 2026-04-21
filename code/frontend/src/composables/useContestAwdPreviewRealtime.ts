import { useWebSocket } from '@/composables/useWebSocket'

export interface ContestAwdPreviewProgressEvent {
  contest_id?: string | number
  preview_request_id?: string
  phase_key?: string
  phase_label?: string
  detail?: string
  attempt?: number
  total_attempts?: number
  status?: string
  error?: string
}

export function useContestAwdPreviewRealtime(
  contestId: string,
  onProgress: (payload: ContestAwdPreviewProgressEvent) => void
) {
  const { status, connect, disconnect } = useWebSocket(`contests/${contestId}/awd-preview`, {
    'awd.preview.progress': (payload) => {
      if (!payload || typeof payload !== 'object') {
        return
      }
      onProgress(payload as ContestAwdPreviewProgressEvent)
    },
  })

  return {
    status,
    start: connect,
    stop: disconnect,
  }
}
