import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import { useAuthStore } from '@/stores/auth'
import {
  contestApiMocks,
  contestDetailPageSource,
  contestDetailRoutePageSource,
  contestDetailSource,
  contestDetailWorkspaceSource,
  contestPresentationSource,
  contestTeamDialogsSource,
  contestTeamPanelSource,
  contestTeamWorkspaceSectionSource,
  destructiveConfirmMock,
  resetContestDetailTestHarness,
  routeQueryTransportSource,
  router,
  teamPresentationSource,
  webSocketMocks,
} from './ContestDetail.test-harness'
import ContestDetail from '@/pages/contests/ContestDetailRoutePage.vue'

describe('ContestDetail', () => {
  beforeEach(async () => {
    await resetContestDetailTestHarness()
  })

  it('不应该向学生暴露草稿竞赛详情或报名入口', async () => {
    contestApiMocks.getContestDetail.mockResolvedValueOnce({
      id: '1',
      title: '草稿竞赛',
      description: '未开放',
      status: 'draft',
      mode: 'jeopardy',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('当前竞赛暂未开放')
    expect(wrapper.text()).not.toContain('创建队伍')
    expect(wrapper.text()).not.toContain('加入队伍')
    expect(wrapper.text()).not.toContain('草稿')
  })

  it('收到公告实时事件后会刷新公告列表', async () => {
    contestApiMocks.getAnnouncements
      .mockResolvedValueOnce([
        {
          id: 'ann-1',
          title: '比赛开始',
          content: '欢迎来到比赛。',
          created_at: '2024-03-15T09:00:00Z',
        },
      ])
      .mockResolvedValueOnce([
        {
          id: 'ann-1',
          title: '比赛开始',
          content: '欢迎来到比赛。',
          created_at: '2024-03-15T09:00:00Z',
        },
        {
          id: 'ann-2',
          title: '第二条公告',
          content: '新的公告已发布。',
          created_at: '2024-03-15T10:00:00Z',
        },
      ])

    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('比赛开始')

    webSocketMocks
      .getHandlers('contests/1/announcements')
      ?.['contest.announcement.created']?.({ contest_id: '1' })

    await flushPromises()

    expect(contestApiMocks.getAnnouncements).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('第二条公告')
  })
})
