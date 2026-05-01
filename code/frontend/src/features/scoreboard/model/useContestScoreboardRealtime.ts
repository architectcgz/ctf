import { useWebSocket } from '@/composables/useWebSocket'
import { useToast } from '@/composables/useToast'

export function useContestScoreboardRealtime(contestId: string, onUpdated: () => void) {
  const toast = useToast()
  const { status, connect, disconnect } = useWebSocket(`contests/${contestId}/scoreboard`, {
    'scoreboard.updated': () => {
      onUpdated()
    },
  })

  async function start(): Promise<void> {
    try {
      await connect()
    } catch (error) {
      toast.warning('实时排行榜连接失败，已切换为手动刷新')
    }
  }

  return {
    status,
    start,
    stop: disconnect,
  }
}
