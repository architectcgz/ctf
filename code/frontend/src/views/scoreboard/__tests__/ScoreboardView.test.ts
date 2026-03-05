import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ScoreboardView from '../ScoreboardView.vue'

vi.mock('@/api/contest', () => ({
  getScoreboard: vi.fn().mockResolvedValue({
    contest: {
      id: 'contest-1',
      title: '测试竞赛',
      status: 'running',
      started_at: '2026-03-05T00:00:00Z',
      ends_at: '2026-03-05T12:00:00Z',
    },
    scoreboard: {
      list: [
        { rank: 1, team_id: 'team-1', team_name: 'Binary Wizards', score: 2450, solved_count: 8 },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    },
    frozen: false,
  }),
}))

describe('ScoreboardView', () => {
  it('应该渲染排行榜页面', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/scoreboard', component: ScoreboardView }],
    })
    await router.push('/scoreboard?contestId=contest-1')
    await router.isReady()

    const wrapper = mount(ScoreboardView, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 50))

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('排行榜')
    expect(wrapper.text()).toContain('Binary Wizards')
  })
})
