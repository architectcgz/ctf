import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'
import { computed, defineComponent, ref } from 'vue'

import type { NotificationItem } from '@/api/contracts'
import type { NotificationDrawerController } from '@/shared/model/layout/notificationDrawerController'
import NotificationDrawer from '../NotificationDrawer.vue'
import { useNotificationStore } from '@/stores/notification'

const notificationApiMocks = vi.hoisted(() => ({
  markAsRead: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  error: vi.fn(),
  warning: vi.fn(),
}))

vi.mock('@/api/notification', () => notificationApiMocks)
vi.mock('@/shared/model/common/useToast', () => ({
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

function mockNotificationDrawerController(router?: ReturnType<typeof createRouter>): NotificationDrawerController {
  const store = useNotificationStore()
  const open = ref(false)
  const isMarkingAllRead = ref(false)

  async function markAllRead() {
    if (isMarkingAllRead.value) return
    const unreadItems = store.notifications.filter((item) => item.unread)
    if (unreadItems.length === 0) return
    isMarkingAllRead.value = true
    try {
      const results = await Promise.allSettled(unreadItems.map((item) => notificationApiMocks.markAsRead(item.id)))
      unreadItems.forEach((item, index) => {
        if (results[index]?.status === 'fulfilled') store.markAsRead(item.id)
      })
      const failedCount = results.filter((r) => r.status === 'rejected').length
      if (failedCount === 0) store.markAllRead()
    } finally {
      isMarkingAllRead.value = false
    }
  }

  return {
    open,
    setTriggerRef: vi.fn(),
    unreadCount: computed(() => store.unreadCount),
    isMarkingAllRead,
    items: computed(() => store.notifications as NotificationItem[]),
    typeMeta: vi.fn().mockReturnValue({ icon: { name: 'Info' }, label: '系统', accentColor: '#000' }),
    close: () => { open.value = false },
    toggleOpen: () => { open.value = !open.value },
    goToNotifications: () => {
      open.value = false
      void router?.push('/notifications')
    },
    goToNotificationDetail: (id: string) => {
      open.value = false
      void router?.push(`/notifications/${encodeURIComponent(id)}`)
    },
    markAllRead,
  }
}

const NotificationDrawerSlotHost = defineComponent({
  components: {
    NotificationDrawer,
  },
  props: {
    controller: {
      type: Object as () => NotificationDrawerController,
      required: true,
    },
  },
  template: `
    <NotificationDrawer realtime-status="open" :controller="controller">
      <template #trigger="{ open, toggle, unreadBadgeLabel, setTriggerRef }">
        <button
          :ref="setTriggerRef"
          type="button"
          class="custom-notification-trigger"
          :aria-expanded="open ? 'true' : 'false'"
          @click="toggle"
        >
          {{ unreadBadgeLabel }}
        </button>
      </template>
    </NotificationDrawer>
  `,
})

async function openDrawer() {
  const router = createTestRouter()
  await router.push('/notifications')
  await router.isReady()

  const wrapper = mount(NotificationDrawer, {
    attachTo: document.body,
    props: {
      realtimeStatus: 'open',
      controller: mockNotificationDrawerController(router),
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

async function openDrawerWithCustomTrigger() {
  const router = createTestRouter()
  await router.push('/notifications')
  await router.isReady()

  const wrapper = mount(NotificationDrawerSlotHost, {
    attachTo: document.body,
    props: {
      controller: mockNotificationDrawerController(router),
    },
    global: {
      plugins: [router],
    },
  })

  await wrapper.get('.custom-notification-trigger').trigger('click')
  await flushPromises()

  return { wrapper, router }
}

describe('NotificationDrawer', () => {
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

  it('supports a custom trigger slot so navigation can own the button shell', async () => {
    const { wrapper } = await openDrawerWithCustomTrigger()

    expect(wrapper.find('.notification-drawer-trigger').exists()).toBe(false)
    expect(wrapper.get('.custom-notification-trigger').text()).toContain('2')
    expect(document.body.textContent).toContain('通知中心')

    wrapper.unmount()
  })

  it('does not render realtime connection status text in the drawer header', async () => {
    const { wrapper } = await openDrawer()

    expect(document.body.textContent).not.toContain('实时同步')
    expect(document.body.textContent).toContain('全部设为已读')

    wrapper.unmount()
  })

  it('owns notification drawer shell dismissal and body scroll lock locally', async () => {
    const { wrapper } = await openDrawer()

    expect(document.body.style.overflow).toBe('hidden')
    expect(document.body.querySelector('.notification-panel')).toBeTruthy()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()

    expect(document.body.querySelector('.notification-panel')).toBeNull()
    expect(document.body.style.overflow).toBe('')

    await wrapper.get('button[aria-label="打开通知中心"]').trigger('click')
    await flushPromises()

    const shell = document.body.querySelector('.notification-shell')
    expect(shell).toBeTruthy()

    shell!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()

    expect(document.body.querySelector('.notification-panel')).toBeNull()
    expect(document.body.style.overflow).toBe('')

    wrapper.unmount()
  })

  it('supports switching between all, unread, and read notification filters', async () => {
    const { wrapper } = await openDrawer()
    const store = useNotificationStore()

    store.markAsRead('1')
    await flushPromises()

    const unreadFilter = Array.from(document.body.querySelectorAll('button')).find(
      (node) => node.textContent?.trim() === '未读'
    )
    const readFilter = Array.from(document.body.querySelectorAll('button')).find(
      (node) => node.textContent?.trim() === '已读'
    )

    expect(unreadFilter).toBeTruthy()
    expect(readFilter).toBeTruthy()

    unreadFilter!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()
    expect(document.body.textContent).toContain('赛事通知')
    expect(document.body.textContent).not.toContain('系统升级公告')

    readFilter!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()
    expect(document.body.textContent).toContain('系统升级公告')
    expect(document.body.textContent).not.toContain('赛事通知')

    wrapper.unmount()
  })

  it('navigates to notification detail when clicking a notification row', async () => {
    const { wrapper, router } = await openDrawer()

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
    const { wrapper, router } = await openDrawer()
    const store = useNotificationStore()

    const markAllButton = Array.from(document.body.querySelectorAll('button')).find((node) =>
      node.textContent?.includes('全部设为已读')
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

  it('prevents duplicate mark-all-read batches while a request is already in flight', async () => {
    notificationApiMocks.markAsRead.mockImplementation(() => new Promise(() => {}))

    const { wrapper } = await openDrawer()
    const markAllButton = Array.from(document.body.querySelectorAll('button')).find((node) =>
      node.textContent?.includes('全部设为已读')
    ) as HTMLButtonElement | undefined

    expect(markAllButton).toBeTruthy()

    markAllButton!.click()
    markAllButton!.click()
    await flushPromises()

    expect(notificationApiMocks.markAsRead).toHaveBeenCalledTimes(2)
    expect(markAllButton!.disabled).toBe(true)

    wrapper.unmount()
  })

  it('restores focus to the trigger after the drawer closes', async () => {
    const { wrapper } = await openDrawer()
    const trigger = wrapper.get('button[aria-label="打开通知中心"]')
    const triggerElement = trigger.element as HTMLButtonElement
    const closeButton = Array.from(document.body.querySelectorAll('button')).find(
      (node) => node.getAttribute('aria-label') === '关闭抽屉'
    )

    triggerElement.focus()
    closeButton!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()

    expect(document.activeElement).toBe(triggerElement)

    wrapper.unmount()
  })
})
