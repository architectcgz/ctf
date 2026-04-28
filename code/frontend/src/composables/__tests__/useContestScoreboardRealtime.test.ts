import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useContestScoreboardRealtime } from '@/composables/useContestScoreboardRealtime'

const toastMocks = vi.hoisted(() => ({
  warning: vi.fn(),
}))

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

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

vi.mock('@/composables/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
}))

describe('useContestScoreboardRealtime', () => {
  beforeEach(() => {
    webSocketMocks.connect.mockClear()
    webSocketMocks.disconnect.mockClear()
    webSocketMocks.send.mockClear()
    webSocketMocks.useWebSocket.mockClear()
    toastMocks.warning.mockReset()
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

  it('连接失败时应在本地消费异常并提示降级为手动刷新', async () => {
    webSocketMocks.connect.mockRejectedValueOnce(new Error('ws unavailable'))

    const { start } = useContestScoreboardRealtime('contest-running', vi.fn())

    await expect(start()).resolves.toBeUndefined()
    expect(webSocketMocks.connect).toHaveBeenCalledTimes(1)
    expect(toastMocks.warning).toHaveBeenCalledWith('实时排行榜连接失败，已切换为手动刷新')
  })
})
