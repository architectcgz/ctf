import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
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
}))

describe('ContestDetail', () => {
  let router: any

  beforeEach(() => {
    router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/contests/:id', component: { template: '<div />' } }],
    })
    router.push('/contests/1')
  })

  it('应该渲染竞赛详情页面', async () => {
    const wrapper = mount(ContestDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.exists()).toBe(true)
  })
})
