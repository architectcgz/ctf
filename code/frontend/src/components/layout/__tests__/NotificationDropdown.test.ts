import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import NotificationDropdown from '../NotificationDropdown.vue'
import { useNotificationStore } from '@/stores/notification'

const notificationApiMocks = vi.hoisted(() => ({
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
      { path: '/notifications', component: { template: '<div>list</div>' } },
      { path: '/notifications/:id', component: { template: '<div>detail</div>' } },
    ],
  })
}

async function openDropdown() {
  const router = createTestRouter()
  await router.push('/notifications')
  await router.isReady()

  const wrapper = mount(NotificationDropdown, {
    attachTo: document.body,
    props: {
      realtimeStatus: 'open',
    },
    global: {
      plugins: [router],
    },
  })

  const trigger = wrapper.find('button[aria-label="打开通知中心"]')
  await trigger.trigger('click')
  await flushPromises()

  return { wrapper, router }
}

describe('NotificationDropdown', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    notificationApiMocks.markAsRead.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    notificationApiMocks.markAsRead.mockResolvedValue(undefined)

    const store = useNotificationStore()
    store.setNotifications([
      {
        id: '1',
        type: 'system',
        title: '系统升级公告',
        content: '升级完成后需要重新登录。',
        unread: true,
        created_at: '2026-03-31T09:00:00Z',
      },
      {
        id: '2',
        type: 'contest',
        title: '赛事通知',
        content: '春季赛题目已发布。',
        unread: true,
        created_at: '2026-03-31T08:00:00Z',
      },
    ])
  })

  it('navigates to notification detail when clicking a timeline item', async () => {
    const { wrapper, router } = await openDropdown()

    const timelineItem = Array.from(document.body.querySelectorAll('button')).find((node) =>
      node.textContent?.includes('系统升级公告')
    )

    expect(timelineItem).toBeTruthy()

    timelineItem!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/notifications/1')
    expect(notificationApiMocks.markAsRead).not.toHaveBeenCalled()

    wrapper.unmount()
  })

  it('keeps mark-all-read and view-all actions available', async () => {
    const { wrapper, router } = await openDropdown()
    const store = useNotificationStore()

    const markAllButton = Array.from(document.body.querySelectorAll('button')).find((node) =>
      node.textContent?.includes('全部标记已读')
    )
    const viewAllButton = Array.from(document.body.querySelectorAll('button')).find((node) =>
      node.textContent?.includes('查看全部通知')
    )

    expect(markAllButton).toBeTruthy()
    expect(viewAllButton).toBeTruthy()

    markAllButton!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()

    expect(notificationApiMocks.markAsRead).toHaveBeenCalledTimes(2)
    expect(store.unreadCount).toBe(0)

    viewAllButton!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/notifications')

    wrapper.unmount()
  })
})
