import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ChallengeList from '../ChallengeList.vue'

vi.mock('@/api/challenge', () => ({
  getChallenges: vi.fn().mockResolvedValue({
    list: [
      {
        id: '1',
        title: 'Test Challenge',
        category: 'web',
        difficulty: 'easy',
        tags: ['test'],
        solved_count: 10,
        total_attempts: 20,
        is_solved: false,
        points: 100,
        created_at: '2024-01-01T00:00:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 12,
  }),
}))

describe('ChallengeList', () => {
  let router: any

  beforeEach(() => {
    router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/challenges/:id', component: { template: '<div />' } }],
    })
  })

  it('应该渲染挑战列表', async () => {
    const wrapper = mount(ChallengeList, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('挑战列表')
  })
})
