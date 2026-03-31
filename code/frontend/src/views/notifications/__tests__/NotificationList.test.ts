import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import NotificationList from '../NotificationList.vue'
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
      { path: '/notifications', component: NotificationList },
      { path: '/notifications/:id', component: { template: '<div>detail</div>' } },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/notifications')
  await router.isReady()

  const wrapper = mount(NotificationList, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('NotificationList', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    notificationApiMocks.getNotifications.mockReset()
    notificationApiMocks.markAsRead.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    notificationApiMocks.markAsRead.mockResolvedValue(undefined)
    notificationApiMocks.getNotifications.mockResolvedValue({
      list: [
        {
          id: '1',
          type: 'system',
          title: '系统通知',
          content: '请及时查看系统更新说明。',
          unread: true,
          created_at: '2026-03-31T09:00:00Z',
        },
        {
          id: '2',
          type: 'contest',
          title: '竞赛通知',
          content: '报名通道已开启。',
          unread: false,
          created_at: '2026-03-31T08:00:00Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
  })

  it('navigates to detail page when clicking a notification item', async () => {
    const { wrapper, router } = await mountPage()

    const notificationButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('系统通知'))

    expect(notificationButton).toBeTruthy()

    await notificationButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/notifications/1')
    expect(notificationApiMocks.markAsRead).not.toHaveBeenCalled()
  })

  it('keeps bulk mark-as-read action working on the list page', async () => {
    const { wrapper } = await mountPage()
    const store = useNotificationStore()

    const bulkReadButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('本页已读'))

    expect(bulkReadButton).toBeTruthy()

    await bulkReadButton!.trigger('click')
    await flushPromises()

    expect(notificationApiMocks.markAsRead).toHaveBeenCalledWith('1')
    expect(store.unreadCount).toBe(0)
  })
})
