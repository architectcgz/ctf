import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ChallengeManage from '../ChallengeManage.vue'

vi.mock('@/api/admin', () => ({
  getChallenges: vi.fn().mockResolvedValue({
    list: [
      {
        id: '1',
        title: 'Test Challenge',
        category: 'web',
        difficulty: 'easy',
        status: 'active',
        base_score: 100,
        solve_count: 5,
        created_at: '2024-01-01T00:00:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 20,
  }),
  createChallenge: vi.fn(),
  updateChallenge: vi.fn(),
  deleteChallenge: vi.fn(),
}))

describe('ChallengeManage', () => {
  it('应该渲染挑战管理页面', async () => {
    const wrapper = mount(ChallengeManage, {
      global: {
        stubs: {
          ElTable: true,
          ElTableColumn: true,
          ElButton: true,
          ElPagination: true,
          ElDialog: true,
          ElForm: true,
          ElFormItem: true,
          ElInput: true,
          ElInputNumber: true,
          ElSelect: true,
          ElOption: true,
        },
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.exists()).toBe(true)
  })
})
