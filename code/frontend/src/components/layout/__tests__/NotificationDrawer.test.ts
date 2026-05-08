import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent } from 'vue'

import NotificationDrawer from '../NotificationDrawer.vue'
import notificationDrawerSource from '../NotificationDrawer.vue?raw'
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

const NotificationDrawerSlotHost = defineComponent({
  components: {
    NotificationDrawer,
  },
  template: `
    <NotificationDrawer realtime-status="open">
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

  it('uses the notification spec drawer shell with compact header chrome', () => {
    expect(notificationDrawerSource).toContain('SlideOverDrawer')
    expect(notificationDrawerSource).toContain('class="notification-shell"')
    expect(notificationDrawerSource).toContain('title="通知中心"')
    expect(notificationDrawerSource).toContain('width="26.875rem"')
    expect(notificationDrawerSource).toContain('body-padding="var(--space-0)"')
    expect(notificationDrawerSource).toContain('<slot')
    expect(notificationDrawerSource).toContain('name="trigger"')
    expect(notificationDrawerSource).toContain(':set-trigger-ref="setTriggerRef"')
    expect(notificationDrawerSource).toContain('<template #header-extra>')
    expect(notificationDrawerSource).toContain('<template #footer>')
    expect(notificationDrawerSource).toContain('全部标为已读')
    expect(notificationDrawerSource).toContain('查看全部通知')
    expect(notificationDrawerSource).toContain('--modal-template-drawer-panel-border')
    expect(notificationDrawerSource).toContain('--modal-template-drawer-header-padding-inline')
    expect(notificationDrawerSource).not.toContain('eyebrow="Notification Hub"')
    expect(notificationDrawerSource).not.toContain(
      'subtitle="集中查看系统、竞赛与训练相关提醒。"'
    )
  })

  it('tokenizes drawer surfaces and keeps header compact around counts, filters, and bottom action', () => {
    expect(notificationDrawerSource).toContain('--notification-surface')
    expect(notificationDrawerSource).toContain('--notification-line')
    expect(notificationDrawerSource).toContain('class="notification-drawer-overview"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-summary"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-summary__main"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-counts"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-status"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-filters"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-filter"')
    expect(notificationDrawerSource).toContain('class="notification-view-all"')
    expect(notificationDrawerSource).toContain('--modal-template-shell-overlay')
    expect(notificationDrawerSource).toContain('--modal-template-drawer-header-surface')
    expect(notificationDrawerSource).toContain('--modal-template-drawer-close-border')
    expect(notificationDrawerSource).toContain('.notification-drawer-filter:hover')
    expect(notificationDrawerSource).toContain('outline: var(--ui-focus-ring-width) solid')
    expect(notificationDrawerSource).not.toContain(':global(.notification-shell')
  })

  it('notification redesign should avoid arbitrary tailwind literals in the drawer chrome', () => {
    expect(notificationDrawerSource).not.toContain('text-[12px]')
    expect(notificationDrawerSource).not.toContain('text-[10px]')
    expect(notificationDrawerSource).not.toContain('w-[1px]')
    expect(notificationDrawerSource).not.toContain('h-[1px]')
  })

  it('通知头部应使用紧凑统计和筛选 pills，并保留补充动作入口', () => {
    expect(notificationDrawerSource).toContain('class="notification-drawer-counts__value"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-counts__total"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-status"')
    expect(notificationDrawerSource).toContain('class="notification-drawer-summary__action"')
    expect(notificationDrawerSource).toContain(
      "'notification-drawer-filter--active': activeFilter === filter.value"
    )
    expect(notificationDrawerSource).toContain("label: '全部'")
    expect(notificationDrawerSource).toContain("label: '未读'")
    expect(notificationDrawerSource).toContain("label: '已读'")
    expect(notificationDrawerSource).toContain('未读')
    expect(notificationDrawerSource).toContain('总计')
    expect(notificationDrawerSource).not.toContain('notification-connection__dot')
    expect(notificationDrawerSource).not.toContain('notification-toolbar__divider')
    expect(notificationDrawerSource).not.toContain('!important')
    expect(notificationDrawerSource).not.toContain('.modal-template-drawer__head-row')
    expect(notificationDrawerSource).not.toContain('class="notification-summary"')
    expect(notificationDrawerSource).not.toContain('padding-inline: var(--space-2);')
  })

  it('通知列表应重构为整行可点击卡片，移除冗余详情按钮与旧时间轴痕迹', () => {
    expect(notificationDrawerSource).toContain('class="notification-list"')
    expect(notificationDrawerSource).toContain('class="notification-item-icon"')
    expect(notificationDrawerSource).toContain('class="notification-item-main"')
    expect(notificationDrawerSource).toContain('class="notification-item-meta"')
    expect(notificationDrawerSource).toContain('class="notification-item-title-row"')
    expect(notificationDrawerSource).toContain('class="notification-item-snippet"')
    expect(notificationDrawerSource).toContain('class="notification-item-unread-dot"')
    expect(notificationDrawerSource).toContain(
      'grid-template-columns: calc(var(--space-5) + var(--space-4)) minmax(0, 1fr) var(--space-2);'
    )
    expect(notificationDrawerSource).toContain('gap: var(--space-4);')
    expect(notificationDrawerSource).toContain(
      'padding: var(--space-4) var(--space-2-5) var(--space-4) var(--space-1-5);'
    )
    expect(notificationDrawerSource).toContain('.notification-panel-body')
    expect(notificationDrawerSource).toContain('border-top: 1px solid var(--notification-line);')
    expect(notificationDrawerSource).toContain('font-size: var(--font-size-1-00);')
    expect(notificationDrawerSource).toContain('font-size: var(--font-size-14);')
    expect(notificationDrawerSource).toContain('white-space: nowrap;')
    expect(notificationDrawerSource).not.toContain(
      'margin-top: calc(var(--space-7) + var(--space-0-5));'
    )
    expect(notificationDrawerSource).not.toContain('查看详情')
    expect(notificationDrawerSource).not.toContain('notification-rail')
    expect(notificationDrawerSource).not.toContain('notification-endcap')
  })

  it('supports a custom trigger slot so navigation can own the button shell', async () => {
    const { wrapper } = await openDrawerWithCustomTrigger()

    expect(wrapper.find('.notification-drawer-trigger').exists()).toBe(false)
    expect(wrapper.get('.custom-notification-trigger').text()).toContain('2')
    expect(document.body.textContent).toContain('通知中心')

    wrapper.unmount()
  })

  it('renders realtime connection status in the drawer header', async () => {
    const { wrapper } = await openDrawer()

    expect(document.body.textContent).toContain('实时在线')

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

    const timelineItem = document.body.querySelector('.notification-item')

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
      node.textContent?.includes('全部标为已读')
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
      node.textContent?.includes('全部标为已读')
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
