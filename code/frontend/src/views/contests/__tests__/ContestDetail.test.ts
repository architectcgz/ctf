import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { createPinia } from 'pinia'
import ContestDetail from '../ContestDetail.vue'

vi.mock('@/api/contest', () => ({
  getContestDetail: vi.fn().mockResolvedValue({
    id: '1',
    title: '2026 春季校园 CTF 挑战赛',
    description: '测试描述',
    status: 'running',
    mode: 'jeopardy',
    starts_at: '2024-03-15T09:00:00Z',
    ends_at: '2024-03-15T21:00:00Z',
  }),
  getMyTeam: vi.fn().mockResolvedValue(null),
  getContestChallenges: vi.fn().mockResolvedValue([]),
  getAnnouncements: vi.fn().mockResolvedValue([
    {
      id: 'ann-1',
      title: '比赛开始',
      content: '欢迎来到比赛。',
      created_at: '2024-03-15T09:00:00Z',
    },
  ]),
  createTeam: vi.fn(),
  joinTeam: vi.fn(),
  kickTeamMember: vi.fn(),
  submitContestFlag: vi.fn(),
}))

describe('ContestDetail', () => {
  let router: any

  beforeEach(async () => {
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

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('公告')
    expect(wrapper.text()).toContain('比赛开始')
  })
})
