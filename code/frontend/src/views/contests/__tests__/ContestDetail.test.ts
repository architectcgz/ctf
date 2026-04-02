import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createPinia } from 'pinia'
import ContestDetail from '../ContestDetail.vue'

const contestApiMocks = vi.hoisted(() => ({
  getContestDetail: vi.fn(),
  getMyTeam: vi.fn(),
  getContestChallenges: vi.fn(),
  getAnnouncements: vi.fn(),
  createTeam: vi.fn(),
  joinTeam: vi.fn(),
  kickTeamMember: vi.fn(),
  submitContestFlag: vi.fn(),
}))

const webSocketMocks = vi.hoisted(() => {
  const connect = vi.fn().mockResolvedValue(undefined)
  const disconnect = vi.fn()
  const handlersByEndpoint = new Map<string, Record<string, (payload: unknown) => void>>()

  return {
    connect,
    disconnect,
    getHandlers: (endpoint: string) => handlersByEndpoint.get(endpoint),
    reset: () => handlersByEndpoint.clear(),
    useWebSocket: vi.fn(
      (endpoint: string, handlers: Record<string, (payload: unknown) => void>) => {
        handlersByEndpoint.set(endpoint, handlers)
        return {
          status: { value: 'idle' as const },
          connect,
          disconnect,
          send: vi.fn(),
        }
      }
    ),
  }
})

vi.mock('@/api/contest', () => contestApiMocks)
vi.mock('@/composables/useWebSocket', () => ({
  useWebSocket: webSocketMocks.useWebSocket,
}))

describe('ContestDetail', () => {
  let router: any

  beforeEach(async () => {
    contestApiMocks.getContestDetail.mockReset()
    contestApiMocks.getMyTeam.mockReset()
    contestApiMocks.getContestChallenges.mockReset()
    contestApiMocks.getAnnouncements.mockReset()
    contestApiMocks.createTeam.mockReset()
    contestApiMocks.joinTeam.mockReset()
    contestApiMocks.kickTeamMember.mockReset()
    contestApiMocks.submitContestFlag.mockReset()
    webSocketMocks.connect.mockClear()
    webSocketMocks.disconnect.mockClear()
    webSocketMocks.useWebSocket.mockClear()
    webSocketMocks.reset()

    contestApiMocks.getContestDetail.mockResolvedValue({
      id: '1',
      title: '2026 春季校园 CTF 挑战赛',
      description: '测试描述',
      status: 'running',
      mode: 'jeopardy',
      starts_at: '2024-03-15T09:00:00Z',
      ends_at: '2024-03-15T21:00:00Z',
    })
    contestApiMocks.getMyTeam.mockResolvedValue(null)
    contestApiMocks.getContestChallenges.mockResolvedValue([])
    contestApiMocks.getAnnouncements.mockResolvedValue([
      {
        id: 'ann-1',
        title: '比赛开始',
        content: '欢迎来到比赛。',
        created_at: '2024-03-15T09:00:00Z',
      },
    ])

    router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/contests/:id', component: { template: '<div />' } }],
    })
    await router.push('/contests/1')
    await router.isReady()
  })

  it('应该渲染竞赛详情页面', async () => {
    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [createPinia(), router],
      },
    })

    await flushPromises()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('公告')
    expect(wrapper.text()).toContain('比赛开始')
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
