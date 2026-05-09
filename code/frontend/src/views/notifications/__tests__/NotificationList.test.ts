import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import NotificationList from '../NotificationList.vue'
import notificationListSource from '../NotificationList.vue?raw'
import notificationCategoryFilterSource from '@/components/notifications/NotificationCategoryFilter.vue?raw'
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

  it('路由页应仅负责组合，不直接耦合通知接口与分页流程', () => {
    expect(notificationListSource).toContain('useNotificationListPage')
    expect(notificationListSource).not.toContain("from '@/api/notification'")
    expect(notificationListSource).not.toContain('usePagination<NotificationItem>')
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

  it('通知页头部应将消息数与未读数收进同一行，并把操作按钮放到下一行', () => {
    expect(notificationListSource).toMatch(
      /<div class="workspace-overline">\s*Notifications\s*<\/div>/
    )
    expect(notificationListSource).toMatch(
      /<h1 class="notification-title workspace-page-title">\s*通知中心\s*<\/h1>/
    )
    expect(notificationListSource).not.toContain('<div class="journal-eyebrow">Notifications</div>')
    expect(notificationListSource).not.toContain('journal-eyebrow-text')
    expect(notificationListSource).toContain('class="notification-topbar-meta"')
    expect(notificationListSource).toContain('class="notification-title-line"')
    expect(notificationListSource).toContain('class="notification-head-stats"')
    expect(notificationListSource).toContain('class="notification-head-stat"')
    expect(notificationListSource).toContain('v-for="stat in headStats"')
    expect(notificationListSource).toContain('{{ stat.label }}')
    expect(notificationListSource).not.toContain('当前消息概况')
    expect(notificationListSource).not.toContain('本页消息')
    expect(notificationListSource).not.toContain('已读消息')
    expect(notificationListSource).not.toContain('总消息数')
  })

  it('通知分类筛选应复用目录工具栏并透传 type 查询参数', async () => {
    const { wrapper } = await mountPage()

    expect(notificationListSource).toContain('NotificationCategoryFilter')
    expect(notificationCategoryFilterSource).toContain('WorkspaceDirectoryToolbar')
    expect(notificationListSource).not.toContain('notification-category-bar')
    expect(wrapper.text()).toContain('全部消息')

    await wrapper.get('.workspace-directory-toolbar__filter-toggle').trigger('click')
    await wrapper.get('.notification-filter-control').setValue('contest')
    await flushPromises()

    expect(wrapper.text()).toContain('竞赛消息')
    expect(notificationApiMocks.getNotifications).toHaveBeenLastCalledWith(
      expect.objectContaining({
        page: 1,
        type: 'contest',
      })
    )
    expect(notificationListSource).not.toContain('class="notification-directory-meta"')
  })

  it('通知页操作按钮应接入共享 ui-btn 原语', () => {
    expect(notificationListSource).toContain('class="ui-btn ui-btn--primary"')
    expect(notificationListSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(notificationListSource).not.toContain('class="notification-btn')
  })

  it('短时间内连续刷新后应显示试探型提示且仍执行真实刷新', async () => {
    const { wrapper } = await mountPage()

    const refreshButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('刷新'))

    expect(refreshButton).toBeTruthy()
    expect(wrapper.text()).not.toContain('新消息不会因为执念刷新得更快。')

    await refreshButton!.trigger('click')
    await flushPromises()
    await refreshButton!.trigger('click')
    await flushPromises()
    await refreshButton!.trigger('click')
    await flushPromises()

    expect(notificationApiMocks.getNotifications).toHaveBeenCalledTimes(4)
    expect(wrapper.text()).toContain('新消息不会因为执念刷新得更快。')
  })
})
