import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, onMounted } from 'vue'

import { useNotificationRealtime } from '@/composables/useNotificationRealtime'
import { useNotificationStore } from '@/stores/notification'

const notificationApiMocks = vi.hoisted(() => ({
  getNotifications: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  error: vi.fn(),
  warning: vi.fn(),
}))

const webSocketMocks = vi.hoisted(() => {
  const connect = vi.fn().mockResolvedValue(undefined)
  const disconnect = vi.fn()
  const send = vi.fn()
  let handlers: Record<string, (payload: unknown) => void> = {}

  return {
    connect,
    disconnect,
    send,
    getHandlers: () => handlers,
    useWebSocket: vi.fn(
      (_endpoint: string, nextHandlers: Record<string, (payload: unknown) => void>) => {
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

vi.mock('@/api/notification', () => notificationApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))
vi.mock('@/composables/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
}))

describe('useNotificationRealtime', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    notificationApiMocks.getNotifications.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    webSocketMocks.connect.mockClear()
    webSocketMocks.disconnect.mockClear()
    webSocketMocks.send.mockClear()
    webSocketMocks.useWebSocket.mockClear()
    notificationApiMocks.getNotifications.mockResolvedValue({
      list: [
        {
          id: '1',
          type: 'system',
          title: '初始化通知',
          unread: true,
          created_at: '2026-03-07T09:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })
  })

  it('hydrates notifications and connects websocket on start', async () => {
    const Harness = defineComponent({
      setup() {
        const { start } = useNotificationRealtime()
        onMounted(() => {
          void start()
        })
        return () => null
      },
    })

    mount(Harness)
    await flushPromises()

    const store = useNotificationStore()
    expect(notificationApiMocks.getNotifications).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(webSocketMocks.connect).toHaveBeenCalledTimes(1)
    expect(store.notifications[0]?.id).toBe('1')
  })

  it('applies created and read websocket events to the store', async () => {
    const Harness = defineComponent({
      setup() {
        const { start } = useNotificationRealtime()
        onMounted(() => {
          void start()
        })
        return () => null
      },
    })

    mount(Harness)
    await flushPromises()

    const handlers = webSocketMocks.getHandlers()
    const store = useNotificationStore()

    handlers['notification.created']?.({
      id: '2',
      type: 'contest',
      title: '实时通知',
      unread: true,
      created_at: '2026-03-07T10:00:00Z',
    })
    handlers['notification.read']?.({
      id: '2',
      type: 'contest',
      title: '实时通知',
      unread: false,
      created_at: '2026-03-07T10:00:00Z',
    })

    expect(store.notifications[0]?.id).toBe('2')
    expect(store.notifications[0]?.unread).toBe(false)
  })
})
