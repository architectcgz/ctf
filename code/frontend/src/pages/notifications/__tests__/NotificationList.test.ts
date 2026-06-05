import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import NotificationList from '@/pages/notifications/NotificationListRoutePage.vue'
import notificationListSource from '@/pages/notifications/NotificationListRoutePage.vue?raw'
import notificationListPageSource from '@/features/notifications/model/useNotificationListPage.ts?raw'
import notificationCategoryFilterSource from '@/features/notifications/ui/NotificationCategoryFilter.vue?raw'
import notificationListWorkspaceSource from '@/widgets/notification-list-workspace/NotificationListWorkspace.vue?raw'
import notificationPresentationSource from '@/entities/notification/model/presentation.ts?raw'
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
vi.mock('@/shared/model/common/useToast', () => ({
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

    const notificationLink = wrapper
      .findAll('.notification-row')
      .find((node) => node.text().includes('系统通知'))

    expect(notificationLink).toBeTruthy()

    await notificationLink!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/notifications/1')
    expect(notificationApiMocks.markAsRead).not.toHaveBeenCalled()
  })

  it('renders the page surface directly as a section inside the layout main area', async () => {
    const { wrapper } = await mountPage()

    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.text()).toContain('Notifications')
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('min-h-full')

    const firstRow = wrapper.find('.notification-row')
    expect(firstRow.find('.notification-row-title').attributes('title')).toBe('系统通知')
    expect(firstRow.find('.notification-row-copy').attributes('title')).toBe(
      '请及时查看系统更新说明。'
    )
  })

  it('路由页应仅负责组合，不直接耦合通知接口与分页流程', () => {
    expect(notificationListSource).toContain('useNotificationListPage')
    expect(notificationListSource).toContain(
      "import { NotificationListWorkspace } from '@/widgets/notification-list-workspace'"
    )
    expect(notificationListSource).not.toContain("from '@/api/notification'")
    expect(notificationListSource).not.toContain('usePagination<NotificationItem>')
    expect(notificationListSource).not.toContain('NotificationCategoryFilter')
    expect(notificationListSource).not.toContain('workspace-directory-grid-row')
    expect(notificationListPageSource).not.toContain("from 'vue-router'")
    expect(notificationListPageSource).toContain('function notificationDetailRoute')
    expect(notificationListPageSource).toContain(
      "import { getNotificationTypeLabel } from '@/entities/notification'"
    )
    expect(notificationListPageSource).not.toContain('function typeLabel')
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

  it('通知页应将消息数与未读数放到分类筛选同行，并把操作按钮留在页头', () => {
    expect(notificationListWorkspaceSource).toContain('Notifications')
    expect(notificationListWorkspaceSource).toContain('通知中心')
    expect(notificationListWorkspaceSource).not.toContain(
      '<div class="journal-eyebrow">Notifications</div>'
    )
    expect(notificationListWorkspaceSource).not.toContain('journal-eyebrow-text')
    expect(notificationListWorkspaceSource).toContain('NotificationCategoryFilter')
    expect(notificationListWorkspaceSource).toContain('v-for="stat in headStats"')
    expect(notificationListWorkspaceSource).toContain('{{ stat.label }}')
    expect(notificationListWorkspaceSource).not.toContain('当前消息概况')
    expect(notificationListWorkspaceSource).not.toContain('本页消息')
    expect(notificationListWorkspaceSource).not.toContain('已读消息')
    expect(notificationListWorkspaceSource).not.toContain('总消息数')
  })

  it('通知分类筛选应复用学生目录筛选样式并透传 type 查询参数', async () => {
    const { wrapper } = await mountPage()

    expect(notificationListWorkspaceSource).toContain('NotificationCategoryFilter')
    expect(notificationListWorkspaceSource).toContain(
      "import { NotificationReadStatePill, NotificationTypePill } from '@/entities/notification'"
    )
    expect(notificationCategoryFilterSource).toContain('student-directory-filters')
    expect(notificationListWorkspaceSource).not.toContain('notification-category-bar')
    expect(wrapper.text()).toContain('全部消息')

    await wrapper.get('.notification-filter-control').setValue('contest')
    await flushPromises()

    expect(wrapper.text()).toContain('竞赛消息')
    expect(notificationApiMocks.getNotifications).toHaveBeenLastCalledWith(
      expect.objectContaining({
        page: 1,
        type: 'contest',
      })
    )
    expect(notificationPresentationSource).toContain("challenge: '训练'")
  })

  it('短时间内连续刷新后应显示试探型提示且仍执行真实刷新', async () => {
    const { wrapper } = await mountPage()

    const refreshButton = wrapper.findAll('button').find((node) => node.text().includes('刷新'))

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
