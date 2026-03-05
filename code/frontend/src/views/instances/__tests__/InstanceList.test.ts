import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import InstanceList from '../InstanceList.vue'

vi.mock('@/api/instance', () => ({
  getMyInstances: vi.fn().mockResolvedValue([
    {
      id: 'inst-1',
      challenge_id: 'chal-1',
      challenge_title: 'SQL 注入基础',
      category: 'web',
      difficulty: 'easy',
      status: 'running',
      access_url: 'http://example.test',
      flag_type: 'static',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 1,
      created_at: '2026-03-05T00:00:00Z',
    },
  ]),
  destroyInstance: vi.fn(),
  extendInstance: vi.fn(),
}))

describe('InstanceList', () => {
  it('应该渲染实例列表页面', async () => {
    const wrapper = mount(InstanceList, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 50))

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('我的实例')
    expect(wrapper.text()).toContain('SQL 注入基础')
  })
})
