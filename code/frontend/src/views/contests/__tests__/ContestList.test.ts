import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ContestList from '../ContestList.vue'

const pushMock = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/contest', () => ({
  getContests: vi.fn().mockResolvedValue({
    list: [
      {
        id: '1',
        title: '2026 春季校园 CTF 挑战赛',
        status: 'running',
        mode: 'jeopardy',
        starts_at: '2024-03-15T09:00:00Z',
        ends_at: '2024-03-15T21:00:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 20,
  }),
}))

describe('ContestList', () => {
  it('应该渲染竞赛列表页面', async () => {
    pushMock.mockReset()

    const wrapper = mount(ContestList)

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('竞赛中心')
  })
})
