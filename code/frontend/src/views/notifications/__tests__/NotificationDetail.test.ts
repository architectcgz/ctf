import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import NotificationDetail from '../NotificationDetail.vue'
import { useNotificationStore } from '@/stores/notification'

const notificationApiMocks = vi.hoisted(() => ({
  getNotifications: vi.fn(),
  markAsRead: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  error: vi.fn(),
  warning: vi.fn(),
}))

vi.mock('@/api/notification', () => notificationApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/notifications', component: { template: '<div />' } },
      { path: '/notifications/:id', component: NotificationDetail },
    ],
  })
}

async function mountPage(path: string) {
  const router = createTestRouter()
  await router.push(path)
  await router.isReady()

  const wrapper = mount(NotificationDetail, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('NotificationDetail', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    notificationApiMocks.getNotifications.mockReset()
    notificationApiMocks.markAsRead.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    notificationApiMocks.getNotifications.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })
    notificationApiMocks.markAsRead.mockResolvedValue(undefined)
  })

  it('renders notification from store and marks unread item as read', async () => {
    const store = useNotificationStore()
    store.setNotifications([
      {
        id: '1',
        type: 'system',
        title: '系统维护窗口',
        content: '今晚 23:00 到 23:30 进行维护。',
        unread: true,
        created_at: '2026-03-31T09:00:00Z',
      },
    ])

    const { wrapper } = await mountPage('/notifications/1')

    expect(wrapper.text()).toContain('系统维护窗口')
    expect(wrapper.text()).toContain('今晚 23:00 到 23:30 进行维护。')
    expect(notificationApiMocks.markAsRead).toHaveBeenCalledWith('1')
    expect(store.notifications[0]?.unread).toBe(false)
  })

  it('falls back to notifications list api when store does not contain the item', async () => {
    notificationApiMocks.getNotifications.mockResolvedValueOnce({
      list: [
        {
          id: '2',
          type: 'contest',
          title: '比赛开始提醒',
          content: '春季赛将在明天 20:00 开始。',
          unread: false,
          created_at: '2026-03-31T08:00:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const { wrapper } = await mountPage('/notifications/2')

    expect(notificationApiMocks.getNotifications).toHaveBeenCalled()
    expect(wrapper.text()).toContain('比赛开始提醒')
    expect(wrapper.text()).toContain('春季赛将在明天 20:00 开始。')
  })

  it('shows empty state when notification cannot be found', async () => {
    const { wrapper } = await mountPage('/notifications/missing')

    expect(wrapper.text()).toContain('通知不存在')
    expect(wrapper.text()).toContain('返回通知列表')
  })
})
