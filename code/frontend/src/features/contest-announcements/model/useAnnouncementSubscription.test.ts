import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAnnouncementSubscription } from '@/features/contest-announcements'

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

vi.mock('@/shared/model/realtime/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
}))

describe('useAnnouncementSubscription', () => {
  beforeEach(() => {
    webSocketMocks.connect.mockClear()
    webSocketMocks.disconnect.mockClear()
    webSocketMocks.useWebSocket.mockClear()
  })

  it('subscribes to contest announcement channel and forwards connect/created/deleted events', async () => {
    const onUpdated = vi.fn()
    const { start, stop } = useAnnouncementSubscription('contest-1', onUpdated)

    await start()

    expect(webSocketMocks.getEndpoint()).toBe('contests/contest-1/announcements')
    webSocketMocks.getHandlers()['system.connected']?.({})
    webSocketMocks.getHandlers()['contest.announcement.created']?.({})
    webSocketMocks.getHandlers()['contest.announcement.deleted']?.({})

    expect(onUpdated).toHaveBeenCalledTimes(3)

    stop()
    expect(webSocketMocks.disconnect).toHaveBeenCalledTimes(1)
  })
})
