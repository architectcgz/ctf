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

  it('uses the reference notification panel shell instead of the generic drawer chrome', () => {
    expect(notificationDrawerSource).toContain('<Teleport to="body">')
    expect(notificationDrawerSource).toContain('class="notification-shell"')
    expect(notificationDrawerSource).toContain('class="notification-panel"')
    expect(notificationDrawerSource).toContain('role="dialog"')
    expect(notificationDrawerSource).toContain('aria-modal="true"')
    expect(notificationDrawerSource).toContain('<slot')
    expect(notificationDrawerSource).toContain('name="trigger"')
    expect(notificationDrawerSource).toContain(':set-trigger-ref="setTriggerRef"')
    expect(notificationDrawerSource).toContain('class="panel-inner"')
    expect(notificationDrawerSource).toContain('class="panel-header"')
    expect(notificationDrawerSource).toContain('NOTIFICATIONS')
    expect(notificationDrawerSource).toContain('class="summary-row"')
    expect(notificationDrawerSource).toContain('class="tabs"')
    expect(notificationDrawerSource).toContain('class="content-divider"')
    expect(notificationDrawerSource).toContain('class="panel-footer"')
    expect(notificationDrawerSource).toContain('全部设为已读')
    expect(notificationDrawerSource).toContain('查看全部通知')
    expect(notificationDrawerSource).not.toContain('ModalTemplateShell')
    expect(notificationDrawerSource).not.toContain('SlideOverDrawer')
    expect(notificationDrawerSource).not.toContain('notificationPanelStyle')
    expect(notificationDrawerSource).not.toContain('--modal-template-drawer-panel-border')
    expect(notificationDrawerSource).not.toContain('系统、竞赛与训练动态按时间更新')
  })

  it('通知面板关键高度与列布局应由通知组件自己的 aside 根节点持有', () => {
    expect(notificationDrawerSource).toContain('.notification-shell {')
    expect(notificationDrawerSource).toContain('position: fixed;')
    expect(notificationDrawerSource).toContain('inset: 0;')
    expect(notificationDrawerSource).toContain('justify-content: flex-end;')
    expect(notificationDrawerSource).toContain('backdrop-filter: blur(7px);')
    expect(notificationDrawerSource).toContain('.notification-panel {')
    expect(notificationDrawerSource).toContain('align-self: stretch;')
    expect(notificationDrawerSource).toContain('display: flex;')
    expect(notificationDrawerSource).toContain('flex-direction: column;')
    expect(notificationDrawerSource).toContain('height: 100dvh;')
    expect(notificationDrawerSource).toContain('min-height: 100dvh;')
    expect(notificationDrawerSource).toContain('overflow: hidden;')
    expect(notificationDrawerSource).toContain(
      '--notification-panel-width: min(40.05vw, 25.3125rem);'
    )
    expect(notificationDrawerSource).toContain('min-width: 23.4375rem;')
    expect(notificationDrawerSource).toContain('max-width: 25.3125rem;')
  })

  it('copies the reference panel rhythm around counts, filters, cards, and bottom action', () => {
    expect(notificationDrawerSource).toContain('class="bell-wrap"')
    expect(notificationDrawerSource).toContain('class="summary-number"')
    expect(notificationDrawerSource).toContain('class="summary-text"')
    expect(notificationDrawerSource).toContain('class="summary-actions"')
    expect(notificationDrawerSource).toContain('class="text-action"')
    expect(notificationDrawerSource).toContain('class="tab-btn"')
    expect(notificationDrawerSource).toContain('class="notice-card"')
    expect(notificationDrawerSource).toContain('class="notice-icon"')
    expect(notificationDrawerSource).toContain('class="notice-title-row"')
    expect(notificationDrawerSource).toContain('class="notice-copy"')
    expect(notificationDrawerSource).toContain('class="unread-dot"')
    expect(notificationDrawerSource).toContain('class="view-all-btn"')
    expect(notificationDrawerSource).toContain('class="footer-icon"')
    expect(notificationDrawerSource).toContain('radial-gradient(')
    expect(notificationDrawerSource).toContain('.tab-btn.is-active')
    expect(notificationDrawerSource).toContain('.view-all-btn')
    expect(notificationDrawerSource).toContain('outline: var(--ui-focus-ring-width) solid')
  })

  it('通知筛选按钮应使用有实体背景的浅深色状态配色', () => {
    const tabsBlock = notificationDrawerSource.match(/\.tabs\s*\{[\s\S]*?\n\}/)?.[0] ?? ''
    const tabButtonBlock = notificationDrawerSource.match(/\.tab-btn\s*\{[\s\S]*?\n\}/)?.[0] ?? ''
    const tabHoverBlock =
      notificationDrawerSource.match(/\.tab-btn:hover:not\(\.is-active\)\s*\{[\s\S]*?\n\}/)?.[0] ??
      ''
    const tabActiveBlock =
      notificationDrawerSource.match(/\.tab-btn\.is-active\s*\{[\s\S]*?\n\}/)?.[0] ?? ''

    expect(notificationDrawerSource).toContain('--notification-tab-shell-bg')
    expect(notificationDrawerSource).toContain('--notification-tab-bg')
    expect(notificationDrawerSource).toContain('--notification-tab-hover-bg')
    expect(notificationDrawerSource).toContain('--notification-tab-active-border')
    expect(notificationDrawerSource).toContain('--notification-tab-active-shadow')
    expect(tabsBlock).toContain('background: var(--notification-tab-shell-bg);')
    expect(tabsBlock).toContain('border: 1px solid var(--notification-tab-shell-border);')
    expect(tabButtonBlock).toContain('background: var(--notification-tab-bg);')
    expect(tabButtonBlock).toContain('border: 1px solid var(--notification-tab-border);')
    expect(tabHoverBlock).toContain('background: var(--notification-tab-hover-bg);')
    expect(tabActiveBlock).toContain('background: var(--notification-tab-active-bg);')
    expect(tabActiveBlock).toContain('border-color: var(--notification-tab-active-border);')
  })

  it('通知抽屉应同时定义白天与夜间模式面板变量', () => {
    expect(notificationDrawerSource).toContain('--notification-panel-surface: rgb(255 255 255);')
    expect(notificationDrawerSource).toContain('--notification-card-bg')
    expect(notificationDrawerSource).toContain('--notification-footer-bg')
    expect(notificationDrawerSource).toContain(":global([data-theme='dark']) .notification-panel")
    expect(notificationDrawerSource).toContain('--notification-panel-surface: rgb(14 23 34);')
    expect(notificationDrawerSource).toContain('--notification-panel-surface-end: rgb(9 18 29);')
    expect(notificationDrawerSource).toContain('background-color: rgb(255 255 255);')
    expect(notificationDrawerSource).toContain(
      'background-image: linear-gradient(180deg, rgb(255 255 255), rgb(244 247 251));'
    )
    expect(notificationDrawerSource).toContain(":global([data-theme='dark']) .panel-inner")
    expect(notificationDrawerSource).toContain('background-color: rgb(14 23 34);')
    expect(notificationDrawerSource).toContain(
      'background-image: linear-gradient(180deg, rgb(14 23 34), rgb(9 18 29));'
    )
    expect(notificationDrawerSource).toContain(
      '--notification-card-bg: linear-gradient(180deg, rgb(255 255 255), rgb(248 250 252));'
    )
    expect(notificationDrawerSource).toContain(
      '--notification-card-bg: linear-gradient(180deg, rgb(18 31 45), rgb(13 23 35));'
    )
    expect(notificationDrawerSource).toContain(
      '--notification-footer-bg: linear-gradient(180deg, rgb(255 255 255), rgb(244 247 251));'
    )
    expect(notificationDrawerSource).toContain(
      '--notification-footer-bg: linear-gradient(180deg, rgb(10 19 30), rgb(8 17 27));'
    )
    expect(notificationDrawerSource).toContain(
      '--notification-panel-shell-bg: linear-gradient(180deg, rgb(255 255 255), rgb(244 247 251));'
    )
    expect(notificationDrawerSource).toContain(
      '--notification-panel-shell-bg: linear-gradient(180deg, rgb(14 23 34), rgb(9 18 29));'
    )
    expect(notificationDrawerSource).toContain(
      'background-image: var(--notification-panel-shell-bg);'
    )
    expect(notificationDrawerSource).not.toContain('rgb(255 255 255 / 0.98)')
    expect(notificationDrawerSource).not.toContain('rgb(244 247 251 / 0.99)')
    expect(notificationDrawerSource).not.toContain('rgb(14 23 34 / 0.98)')
    expect(notificationDrawerSource).not.toContain('rgb(9 18 29 / 0.99)')
  })

  it('通知底部操作应处于正常布局流，避免在内容区上方形成透明悬浮层', () => {
    const panelBlock =
      notificationDrawerSource.match(/\.notification-panel\s*\{[\s\S]*?\n\}/)?.[0] ?? ''
    const footerBlock = notificationDrawerSource.match(/\.panel-footer\s*\{[\s\S]*?\n\}/)?.[0] ?? ''
    const viewAllButtonBlock =
      notificationDrawerSource.match(/\.view-all-btn\s*\{[\s\S]*?\n\}/)?.[0] ?? ''

    expect(panelBlock).toContain('align-self: stretch;')
    expect(panelBlock).toContain('height: 100dvh;')
    expect(panelBlock).toContain('min-height: 100dvh;')
    expect(panelBlock).toContain('display: flex;')
    expect(panelBlock).toContain('flex-direction: column;')
    expect(notificationDrawerSource).not.toContain('grid-template-rows: minmax(0, 1fr) auto;')
    expect(footerBlock).toContain('position: relative;')
    expect(notificationDrawerSource).toContain('flex: 1 1 0;')
    expect(footerBlock).toContain('flex: 0 0 auto;')
    expect(footerBlock).toContain('margin-top: auto;')
    expect(footerBlock).toContain('box-shadow: none;')
    expect(footerBlock).toContain(
      'padding: var(--space-3) 0 calc(var(--space-3) + env(safe-area-inset-bottom, 0px));'
    )
    expect(footerBlock).toContain('background-color: rgb(255 255 255);')
    expect(footerBlock).toContain('background-image: var(--notification-footer-bg);')
    expect(viewAllButtonBlock).toContain('min-height: var(--ui-control-height-lg);')
    expect(viewAllButtonBlock).toContain('background-color: rgb(255 255 255);')
    expect(viewAllButtonBlock).toContain('background-image: var(--notification-footer-bg);')
    expect(footerBlock).not.toContain('min-height: calc(5.75rem')
    expect(viewAllButtonBlock).not.toContain('height: 5.75rem;')
    expect(footerBlock).not.toContain('position: absolute;')
    expect(footerBlock).not.toContain('bottom: 0;')
    expect(footerBlock).not.toContain('0 -1.125rem 2.375rem')
    expect(notificationDrawerSource).not.toContain('padding: 3rem 2rem 6.75rem 1.5rem;')
  })

  it('notification redesign should avoid arbitrary tailwind literals in the drawer chrome', () => {
    expect(notificationDrawerSource).not.toContain('text-[12px]')
    expect(notificationDrawerSource).not.toContain('text-[10px]')
    expect(notificationDrawerSource).not.toContain('w-[1px]')
    expect(notificationDrawerSource).not.toContain('h-[1px]')
  })

  it('通知头部应使用参考稿统计行和筛选按钮，并保留补充动作入口', () => {
    expect(notificationDrawerSource).toContain('class="summary-number"')
    expect(notificationDrawerSource).toContain('class="summary-text"')
    expect(notificationDrawerSource).toContain('class="summary-actions"')
    expect(notificationDrawerSource).toContain('全部设为已读')
    expect(notificationDrawerSource).not.toContain('class="text-action text-action--status"')
    expect(notificationDrawerSource).not.toContain('class="action-separator"')
    expect(notificationDrawerSource).not.toContain('实时同步')
    expect(notificationDrawerSource).toContain("'is-active': activeFilter === filter.value")
    expect(notificationDrawerSource).toContain("label: '全部'")
    expect(notificationDrawerSource).toContain("label: '未读'")
    expect(notificationDrawerSource).toContain("label: '已读'")
    expect(notificationDrawerSource).toContain('条未读通知待处理')
    expect(notificationDrawerSource).toContain('全部通知已读')
    expect(notificationDrawerSource).toContain('当前没有新通知')
    expect(notificationDrawerSource).not.toContain('notification-connection__dot')
    expect(notificationDrawerSource).not.toContain('notification-toolbar__divider')
    expect(notificationDrawerSource).not.toContain('!important')
    expect(notificationDrawerSource).not.toContain('.modal-template-drawer__head-row')
    expect(notificationDrawerSource).not.toContain('class="notification-summary"')
    expect(notificationDrawerSource).not.toContain('class="notification-drawer-filter"')
  })

  it('通知列表应复刻参考稿卡片，移除冗余详情按钮与旧时间轴痕迹', () => {
    expect(notificationDrawerSource).toContain('class="notification-list"')
    expect(notificationDrawerSource).toContain('class="notice-card"')
    expect(notificationDrawerSource).toContain('class="notice-icon"')
    expect(notificationDrawerSource).toContain('class="notice-body"')
    expect(notificationDrawerSource).toContain('class="notice-category"')
    expect(notificationDrawerSource).toContain('class="notice-title-row"')
    expect(notificationDrawerSource).toContain('class="notice-copy"')
    expect(notificationDrawerSource).toContain('class="unread-dot"')
    expect(notificationDrawerSource).toContain('grid-template-columns: 2.75rem minmax(0, 1fr);')
    expect(notificationDrawerSource).toContain('min-height: 5.875rem;')
    expect(notificationDrawerSource).toContain('border-radius: var(--ui-control-radius-lg);')
    expect(notificationDrawerSource).toContain('font-size: var(--font-size-15);')
    expect(notificationDrawerSource).toContain('font-size: var(--font-size-13);')
    expect(notificationDrawerSource).toContain('background: var(--notification-card-bg);')
    expect(notificationDrawerSource).toContain('white-space: nowrap;')
    expect(notificationDrawerSource).not.toContain('class="notification-item"')
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

    const timelineItem = document.body.querySelector('.notice-card')

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
