import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import NotificationDetail from '@/pages/notifications/NotificationDetailRoutePage.vue'
import notificationDetailSource from '@/pages/notifications/NotificationDetailRoutePage.vue?raw'
import notificationDetailPageSource from '@/features/notifications/model/useNotificationDetailPage.ts?raw'
import notificationDetailWorkspaceSource from '@/widgets/notification-detail-workspace/NotificationDetailWorkspace.vue?raw'
import notificationPresentationSource from '@/entities/notification/model/presentation.ts?raw'
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
vi.mock('@/shared/model/common/useToast', () => ({
  useToast: () => toastMocks,
}))

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/notifications', name: 'Notifications', component: { template: '<div />' } },
      { path: '/platform/overview', component: { template: '<div />' } },
      { path: '/challenges/:id', component: { template: '<div />' } },
      { path: '/contests/:id', component: { template: '<div />' } },
      {
        path: '/notifications/:id',
        name: 'NotificationDetail',
        component: NotificationDetail,
        props: (route) => ({ id: String(route.params.id || '') }),
      },
    ],
  })
}

async function mountPage(path: string) {
  const router = createTestRouter()
  await router.push(path)
  await router.isReady()

  const wrapper = mount(NotificationDetail, {
    props: {
      id: String(router.currentRoute.value.params.id || ''),
    },
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
        link: '/platform/overview',
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

  it('路由页应仅负责组合，不直接耦合通知详情读取流程', () => {
    expect(notificationDetailSource).toContain('useNotificationDetailPage')
    expect(notificationDetailSource).toContain("useNotificationDetailPage(toRef(props, 'id'))")
    expect(notificationDetailSource).toContain(
      "import { NotificationDetailWorkspace } from '@/widgets/notification-detail-workspace'"
    )
    expect(notificationDetailSource).not.toContain("from '@/api/notification'")
    expect(notificationDetailSource).not.toContain('useRoute(')
    expect(notificationDetailSource).not.toContain('useRouter(')
    expect(notificationDetailSource).not.toContain('watch(')
    expect(notificationDetailSource).not.toContain('workspace-overline')
    expect(notificationDetailPageSource).not.toContain("from 'vue-router'")
    expect(notificationDetailPageSource).not.toContain('accentColorMap')
    expect(notificationDetailPageSource).not.toContain('function notificationAccent')
    expect(notificationDetailPageSource).not.toContain('function notificationTypeLabel')
  })

  it('falls back to notifications list api when store does not contain the item', async () => {
    notificationApiMocks.getNotifications.mockResolvedValueOnce({
      list: [
        {
          id: '2',
          type: 'contest',
          title: '比赛开始提醒',
          content: '春季赛将在明天 20:00 开始。',
          link: '/contests/2',
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

  it('通知详情页 overline 应接入 workspace-overline 共享语义', () => {
    expect(notificationDetailWorkspaceSource).toContain('Notification')
    expect(notificationDetailWorkspaceSource).toContain('Meta')
    expect(notificationDetailWorkspaceSource).toContain('ID')
    expect(notificationDetailWorkspaceSource).toContain('Message')
    expect(notificationDetailWorkspaceSource).not.toContain(
      '<div class="notification-overline">Notification</div>'
    )
    expect(notificationDetailWorkspaceSource).not.toContain(
      '<div class="notification-overline">Meta</div>'
    )
    expect(notificationDetailWorkspaceSource).not.toContain(
      '<div class="notification-overline">ID</div>'
    )
    expect(notificationDetailWorkspaceSource).not.toContain(
      '<div class="notification-overline">Message</div>'
    )
    expect(notificationDetailWorkspaceSource).not.toMatch(/^\.notification-overline\s*\{/m)
    expect(notificationPresentationSource).toContain("team: 'violet'")
  })

  it('存在站内 link 时应通过 route target 渲染关联入口', async () => {
    const store = useNotificationStore()
    store.setNotifications([
      {
        id: '9',
        type: 'challenge',
        title: '题目更新',
        content: '请查看最新题目说明。',
        link: '/challenges/9',
        unread: false,
        created_at: '2026-03-31T10:00:00Z',
      },
    ])

    const { wrapper } = await mountPage('/notifications/9')

    expect(wrapper.text()).toContain('查看关联对象')
    expect(wrapper.find('button[disabled]').exists()).toBe(false)
    expect(notificationDetailWorkspaceSource).toContain(
      "import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'"
    )
    expect(wrapper.get('a[href="/challenges/9"]').text()).toContain('查看关联对象')
  })

  it('存在外链时应保留新窗口打开入口', async () => {
    const store = useNotificationStore()
    store.setNotifications([
      {
        id: '10',
        type: 'system',
        title: '外部公告',
        content: '查看外部公告详情。',
        link: 'https://example.com/announcements/10',
        unread: false,
        created_at: '2026-03-31T10:05:00Z',
      },
    ])

    const { wrapper } = await mountPage('/notifications/10')

    const relatedLink = wrapper.get('a[href="https://example.com/announcements/10"]')
    expect(relatedLink.attributes('target')).toBe('_blank')
    expect(relatedLink.attributes('rel')).toContain('noopener')
  })

  it('连续点击 ID 卡片后应短暂显示值守备注', async () => {
    const store = useNotificationStore()
    store.setNotifications([
      {
        id: '9',
        type: 'challenge',
        title: '题目更新',
        content: '请查看最新题目说明。',
        link: '/challenges/9',
        unread: false,
        created_at: '2026-03-31T10:00:00Z',
      },
    ])

    const { wrapper } = await mountPage('/notifications/9')

    expect(wrapper.text()).toContain('9')
    expect(wrapper.text()).not.toContain('值守备注：有人开始认真看编号了。')

    const idCard = wrapper.get('.notification-detail-side-value--mono')
    await idCard.trigger('click')
    await idCard.trigger('click')
    await idCard.trigger('click')
    await idCard.trigger('click')

    expect(wrapper.text()).toContain('值守备注：有人开始认真看编号了。')
    expect(wrapper.text()).toContain('请查看最新题目说明。')
    expect(wrapper.text()).toContain('返回通知列表')
  })
})
