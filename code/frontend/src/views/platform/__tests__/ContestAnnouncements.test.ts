import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestAnnouncements from '../ContestAnnouncements.vue'
import contestAnnouncementsSource from '../ContestAnnouncements.vue?raw'
import contestAnnouncementsTopbarPanelSource from '@/components/platform/contest/ContestAnnouncementsTopbarPanel.vue?raw'
import routerSource from '@/router/index.ts?raw'

const pushMock = vi.fn()
const routeState = vi.hoisted(() => ({
  params: { id: 'contest-1' } as Record<string, string>,
}))

const adminApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
  getAdminContestAnnouncements: vi.fn(),
  createAdminContestAnnouncement: vi.fn(),
  deleteAdminContestAnnouncement: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: pushMock, replace: vi.fn(), back: vi.fn() }),
  }
})

vi.mock('@/api/admin/contests', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/contests')>('@/api/admin/contests')
  return {
    ...actual,
    getContest: adminApiMocks.getContest,
    getAdminContestAnnouncements: adminApiMocks.getAdminContestAnnouncements,
    createAdminContestAnnouncement: adminApiMocks.createAdminContestAnnouncement,
    deleteAdminContestAnnouncement: adminApiMocks.deleteAdminContestAnnouncement,
  }
})

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('ContestAnnouncements', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.getContest.mockReset()
    adminApiMocks.getAdminContestAnnouncements.mockReset()
    adminApiMocks.createAdminContestAnnouncement.mockReset()
    adminApiMocks.deleteAdminContestAnnouncement.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()
    toastMocks.info.mockReset()

    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 春季赛',
      description: '公告运营',
      mode: 'jeopardy',
      status: 'running',
      starts_at: '2026-04-22T08:00:00.000Z',
      ends_at: '2026-04-22T18:00:00.000Z',
      scoreboard_frozen: false,
    })
    adminApiMocks.getAdminContestAnnouncements.mockResolvedValue([
      {
        id: 'announcement-1',
        title: '报名提醒',
        content: '请在今晚前完成组队。',
        created_at: '2026-04-22T09:00:00.000Z',
      },
    ])
    adminApiMocks.createAdminContestAnnouncement.mockResolvedValue({
      id: 'announcement-2',
      title: '开赛通知',
      content: '比赛将于十分钟后开始。',
      created_at: '2026-04-22T09:10:00.000Z',
    })
    adminApiMocks.deleteAdminContestAnnouncement.mockResolvedValue(undefined)
  })

  it('应注册单场公告管理路由', () => {
    expect(routerSource).toContain("path: 'platform/contests/:id/announcements'")
    expect(routerSource).toContain("name: 'ContestAnnouncements'")
    expect(routerSource).toContain(
      "component: () => import('@/views/platform/ContestAnnouncements.vue')"
    )
  })

  it('页面应加载竞赛详情和公告列表', async () => {
    expect(contestAnnouncementsTopbarPanelSource).toContain('Contest Announcements')
    expect(contestAnnouncementsTopbarPanelSource).toContain('class="contest-announcement-status"')

    const wrapper = mount(ContestAnnouncements)

    await flushPromises()

    expect(adminApiMocks.getContest).toHaveBeenCalledWith('contest-1')
    expect(adminApiMocks.getAdminContestAnnouncements).toHaveBeenCalledWith('contest-1')
    expect(wrapper.text()).toContain('2026 春季赛')
    expect(wrapper.text()).toContain('报名提醒')
    expect(wrapper.find('#contest-announcement-submit').exists()).toBe(true)
  })

  it('路由页应仅负责组合，不直接耦合公告页加载流程', () => {
    expect(contestAnnouncementsSource).toContain('useContestAnnouncementsPage')
    expect(contestAnnouncementsSource).not.toContain("from '@/api/admin/contests'")
  })

  it('ended 竞赛应显示只读提示，且不显示发布和删除操作', async () => {
    adminApiMocks.getContest.mockResolvedValueOnce({
      id: 'contest-1',
      title: '2026 春季赛',
      description: '公告运营',
      mode: 'jeopardy',
      status: 'ended',
      starts_at: '2026-04-22T08:00:00.000Z',
      ends_at: '2026-04-22T18:00:00.000Z',
      scoreboard_frozen: false,
    })

    const wrapper = mount(ContestAnnouncements)

    await flushPromises()

    expect(wrapper.text()).toContain('赛事已结束，公告区仅保留查看能力。')
    expect(wrapper.find('#contest-announcement-submit').exists()).toBe(false)
    expect(wrapper.find('#contest-announcement-delete-announcement-1').exists()).toBe(false)
  })
})
