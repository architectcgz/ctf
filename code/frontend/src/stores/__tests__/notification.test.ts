import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { useNotificationStore } from '@/stores/notification'

describe('notification store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('dedupes and sorts notifications when hydrating', () => {
    const store = useNotificationStore()

    store.setNotifications([
      {
        id: '2',
        type: 'contest',
        title: '较新的通知',
        unread: true,
        created_at: '2026-03-07T10:00:00Z',
      },
      {
        id: '1',
        type: 'system',
        title: '较早的通知',
        unread: false,
        created_at: '2026-03-07T08:00:00Z',
      },
      {
        id: '2',
        type: 'contest',
        title: '较新的通知',
        unread: true,
        created_at: '2026-03-07T10:00:00Z',
      },
    ])

    expect(store.notifications).toHaveLength(2)
    expect(store.notifications[0].id).toBe('2')
    expect(store.unreadCount).toBe(1)
  })

  it('upserts latest notification and updates read state', () => {
    const store = useNotificationStore()

    store.setNotifications([
      {
        id: '1',
        type: 'system',
        title: '旧通知',
        unread: true,
        created_at: '2026-03-07T08:00:00Z',
      },
    ])

    store.upsertNotification({
      id: '2',
      type: 'challenge',
      title: '新通知',
      unread: true,
      created_at: '2026-03-07T12:00:00Z',
    })
    store.markAsRead('2')

    expect(store.notifications[0].id).toBe('2')
    expect(store.notifications[0].unread).toBe(false)
  })
})
