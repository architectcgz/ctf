import { useWebSocket } from '@/composables/useWebSocket'

export function useContestScoreboardRealtime(contestId: string, onUpdated: () => void) {
  const { status, connect, disconnect } = useWebSocket(`contests/${contestId}/scoreboard`, {
    'scoreboard.updated': () => {
      onUpdated()
    },
  })

  return {
    status,
    start: connect,
    stop: disconnect,
  }
}
