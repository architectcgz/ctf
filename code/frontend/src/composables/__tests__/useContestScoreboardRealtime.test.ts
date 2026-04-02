import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useContestScoreboardRealtime } from '@/composables/useContestScoreboardRealtime'

const webSocketMocks = vi.hoisted(() => {
  const connect = vi.fn().mockResolvedValue(undefined)
  const disconnect = vi.fn()
  const send = vi.fn()
  let endpoint = ''
  let handlers: Record<string, (payload: unknown) => void> = {}

  return {
    connect,
    disconnect,
    send,
    getEndpoint: () => endpoint,
    getHandlers: () => handlers,
    useWebSocket: vi.fn(
      (nextEndpoint: string, nextHandlers: Record<string, (payload: unknown) => void>) => {
        endpoint = nextEndpoint
        handlers = nextHandlers
        return {
          status: { value: 'idle' as const },
          connect,
          disconnect,
          send,
        }
      }
    ),
  }
})

vi.mock('@/composables/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
}))

describe('useContestScoreboardRealtime', () => {
  beforeEach(() => {
    webSocketMocks.connect.mockClear()
    webSocketMocks.disconnect.mockClear()
    webSocketMocks.send.mockClear()
    webSocketMocks.useWebSocket.mockClear()
  })

  it('subscribes to contest scoreboard channel and forwards update events', async () => {
    const onUpdated = vi.fn()
    const { start, stop } = useContestScoreboardRealtime('contest-running', onUpdated)

    await start()

    expect(webSocketMocks.getEndpoint()).toBe('contests/contest-running/scoreboard')
    expect(webSocketMocks.connect).toHaveBeenCalledTimes(1)

    webSocketMocks.getHandlers()['scoreboard.updated']?.({ contest_id: 'contest-running' })
    expect(onUpdated).toHaveBeenCalledTimes(1)

    stop()
    expect(webSocketMocks.disconnect).toHaveBeenCalledTimes(1)
  })
})
