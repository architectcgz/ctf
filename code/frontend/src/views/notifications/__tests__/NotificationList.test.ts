import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import NotificationList from '../NotificationList.vue'
import notificationListSource from '../NotificationList.vue?raw'
import { useNotificationStore } from '@/stores/notification'
import { useAuthStore } from '@/stores/auth'

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

async function mountPage(role: 'student' | 'teacher' | 'admin' = 'student') {
  const router = createTestRouter()
  await router.push('/notifications')
  await router.isReady()
  const authStore = useAuthStore()
  authStore.user = {
    id: 'u-1',
    username: 'tester',
    role,
  }
  authStore.accessToken = 'token'

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

  it('renders the page surface directly as a section inside the layout main area', async () => {
    const { wrapper } = await mountPage()
    const className = wrapper.attributes('class')

    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.text()).toContain('Notifications')
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.classes()).toContain('space-y-6')
    expect(className).not.toContain('-mx-4')
    expect(className).not.toContain('-my-6')
    expect(className).not.toContain('md:-mx-6')
    expect(className).not.toContain('xl:-mx-8')
    expect(className).not.toContain('md:min-h-[calc(100vh-5rem)]')

    const firstRow = wrapper.find('.notification-row')
    expect(firstRow.find('.notification-row-title').attributes('title')).toBe('系统通知')
    expect(firstRow.find('.notification-row-copy').attributes('title')).toBe(
      '请及时查看系统更新说明。'
    )
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

  it('shows publish entry for admin and hides for non-admin users', async () => {
    const adminPage = await mountPage('admin')
    expect(adminPage.wrapper.text()).toContain('发布通知')

    const teacherPage = await mountPage('teacher')
    expect(teacherPage.wrapper.text()).not.toContain('发布通知')
  })

  it('keeps notification list titles and content truncated with full hover text', () => {
    expect(notificationListSource).toMatch(
      /class="notification-row-title"[\s\S]*:title="item\.title"/s
    )
    expect(notificationListSource).toMatch(
      /class="notification-row-copy"[\s\S]*:title="item\.content"/s
    )
    expect(notificationListSource).toMatch(
      /\.notification-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(notificationListSource).toMatch(
      /\.notification-row-copy\s*\{[^}]*display:\s*-webkit-box;[^}]*-webkit-line-clamp:\s*2;[^}]*overflow:\s*hidden;/s
    )
  })

  it('通知页概况卡片应使用统一 metric-panel 样式类', () => {
    expect(notificationListSource).toContain('class="notification-summary-grid metric-panel-grid"')
    expect(notificationListSource).toContain('class="notification-summary-item metric-panel-card"')
    expect(notificationListSource).toContain(
      'class="notification-summary-label metric-panel-label"'
    )
    expect(notificationListSource).toContain(
      'class="notification-summary-value metric-panel-value"'
    )
    expect(notificationListSource).toContain(
      'class="notification-summary-helper metric-panel-helper"'
    )
  })
})
