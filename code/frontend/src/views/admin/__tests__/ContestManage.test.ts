import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ContestManage from '../ContestManage.vue'

const contestMocks = vi.hoisted(() => ({
  getContests: vi.fn(),
  createContest: vi.fn(),
  updateContest: vi.fn(),
}))

vi.mock('@/api/admin', async () => {
  const actual = await vi.importActual<typeof import('@/api/admin')>('@/api/admin')
  return {
    ...actual,
    getContests: contestMocks.getContests,
    createContest: contestMocks.createContest,
    updateContest: contestMocks.updateContest,
  }
})

describe('ContestManage', () => {
  beforeEach(() => {
    contestMocks.getContests.mockReset()
    contestMocks.createContest.mockReset()
    contestMocks.updateContest.mockReset()
  })

  it('应该渲染真实竞赛列表', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [
        {
          id: '1',
          title: '2026 春季校园 CTF',
          description: '校内赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-15T09:00:00.000Z',
          ends_at: '2026-03-15T13:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('赛事编排台')
    expect(wrapper.text()).toContain('2026 春季校园 CTF')
    expect(wrapper.text()).toContain('报名中')
    expect(contestMocks.getContests).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: undefined,
    })
  })

  it('应该在空列表时展示显式空态', async () => {
    contestMocks.getContests.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    const wrapper = mount(ContestManage, {
      global: {
        stubs: {
          ElDialog: {
            template: '<div><slot /><slot name="footer" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('暂无竞赛')
    expect(wrapper.text()).not.toContain('mockContests')
  })
})
