import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import ChallengeDetail from '../ChallengeDetail.vue'

vi.mock('@/api/challenge', () => ({
  getChallengeDetail: vi.fn().mockResolvedValue({
    id: '1',
    title: 'Test Challenge',
    description: '<p>Test description</p>',
    category: 'web',
    difficulty: 'easy',
    tags: ['test'],
    points: 100,
    is_solved: false,
    hints: [],
  }),
  submitFlag: vi.fn(),
  createInstance: vi.fn(),
}))

describe('ChallengeDetail', () => {
  let router: any

  beforeEach(() => {
    router = createRouter({
      history: createMemoryHistory(),
      routes: [{ path: '/challenges/:id', component: { template: '<div />' } }],
    })
    router.push('/challenges/1')
  })

  it('应该渲染挑战详情', async () => {
    const wrapper = mount(ChallengeDetail, {
      global: {
        plugins: [router],
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.text()).toContain('Test Challenge')
  })
})
