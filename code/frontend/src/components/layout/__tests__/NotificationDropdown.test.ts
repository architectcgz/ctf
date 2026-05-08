import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import NotificationDropdown from '../NotificationDropdown.vue'
import notificationDropdownSource from '../NotificationDropdown.vue?raw'
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

  it('uses the notification spec drawer shell with compact header chrome', () => {
    expect(notificationDropdownSource).toContain('SlideOverDrawer')
    expect(notificationDropdownSource).toContain('class="notification-shell"')
    expect(notificationDropdownSource).toContain('title="通知中心"')
    expect(notificationDropdownSource).toContain('width="26.875rem"')
    expect(notificationDropdownSource).toContain('body-padding="var(--space-0)"')
    expect(notificationDropdownSource).toContain('<template #header-extra>')
    expect(notificationDropdownSource).toContain('<template #footer>')
    expect(notificationDropdownSource).toContain('全部标为已读')
    expect(notificationDropdownSource).toContain('查看全部通知')
    expect(notificationDropdownSource).toContain('--modal-template-drawer-panel-border')
    expect(notificationDropdownSource).toContain('--modal-template-drawer-header-padding-inline')
    expect(notificationDropdownSource).not.toContain('eyebrow="Notification Hub"')
    expect(notificationDropdownSource).not.toContain(
      'subtitle="集中查看系统、竞赛与训练相关提醒。"'
    )
  })

  it('tokenizes drawer surfaces and keeps header compact around counts, filters, and bottom action', () => {
    expect(notificationDropdownSource).toContain('--notification-surface')
    expect(notificationDropdownSource).toContain('--notification-line')
    expect(notificationDropdownSource).toContain('class="notification-overview"')
    expect(notificationDropdownSource).toContain('class="notification-summary"')
    expect(notificationDropdownSource).toContain('class="notification-counts"')
    expect(notificationDropdownSource).toContain('class="notification-filter-tabs"')
    expect(notificationDropdownSource).toContain('class="notification-filter"')
    expect(notificationDropdownSource).toContain('class="notification-view-all"')
    expect(notificationDropdownSource).toContain('--modal-template-shell-overlay')
    expect(notificationDropdownSource).toContain('--modal-template-drawer-header-surface')
    expect(notificationDropdownSource).toContain('--modal-template-drawer-close-border')
    expect(notificationDropdownSource).toContain('.notification-filter:hover')
    expect(notificationDropdownSource).toContain('outline: var(--ui-focus-ring-width) solid')
    expect(notificationDropdownSource).not.toContain(':global(.notification-shell')
  })

  it('notification redesign should avoid arbitrary tailwind literals in the drawer chrome', () => {
    expect(notificationDropdownSource).not.toContain('text-[12px]')
    expect(notificationDropdownSource).not.toContain('text-[10px]')
    expect(notificationDropdownSource).not.toContain('w-[1px]')
    expect(notificationDropdownSource).not.toContain('h-[1px]')
  })

  it('通知头部应使用紧凑统计和筛选 pills，并保留补充动作入口', () => {
    expect(notificationDropdownSource).toContain('class="notification-counts__value"')
    expect(notificationDropdownSource).toContain('class="notification-counts__total"')
    expect(notificationDropdownSource).toContain('class="notification-summary__action"')
    expect(notificationDropdownSource).toContain(
      "'notification-filter--active': activeFilter === filter.value"
    )
    expect(notificationDropdownSource).toContain("label: '全部'")
    expect(notificationDropdownSource).toContain("label: '未读'")
    expect(notificationDropdownSource).toContain("label: '已读'")
    expect(notificationDropdownSource).toContain('未读')
    expect(notificationDropdownSource).toContain('总计')
    expect(notificationDropdownSource).not.toContain('notification-connection__dot')
    expect(notificationDropdownSource).not.toContain('notification-toolbar__divider')
  })

  it('通知列表应重构为整行可点击卡片，移除冗余详情按钮与旧时间轴痕迹', () => {
    expect(notificationDropdownSource).toContain('class="notification-list"')
    expect(notificationDropdownSource).toContain('class="notification-item-icon"')
    expect(notificationDropdownSource).toContain('class="notification-item-main"')
    expect(notificationDropdownSource).toContain('class="notification-item-meta"')
    expect(notificationDropdownSource).toContain('class="notification-item-title-row"')
    expect(notificationDropdownSource).toContain('class="notification-item-snippet"')
    expect(notificationDropdownSource).toContain('class="notification-item-unread-dot"')
    expect(notificationDropdownSource).toContain(
      'grid-template-columns: calc(var(--space-5) + var(--space-4)) minmax(0, 1fr) var(--space-2);'
    )
    expect(notificationDropdownSource).toContain('gap: var(--space-4);')
    expect(notificationDropdownSource).toContain(
      'padding: var(--space-4) var(--space-2-5) var(--space-4) var(--space-1-5);'
    )
    expect(notificationDropdownSource).toContain('font-size: var(--font-size-1-00);')
    expect(notificationDropdownSource).toContain('font-size: var(--font-size-14);')
    expect(notificationDropdownSource).toContain('white-space: nowrap;')
    expect(notificationDropdownSource).not.toContain('查看详情')
    expect(notificationDropdownSource).not.toContain('notification-rail')
    expect(notificationDropdownSource).not.toContain('notification-endcap')
  })

  it('supports switching between all, unread, and read notification filters', async () => {
    const { wrapper } = await openDropdown()
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
    const { wrapper, router } = await openDropdown()

    const timelineItem = document.body.querySelector('.notification-item')

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
})
