import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import InstanceList from '../InstanceList.vue'

const instanceApiMocks = vi.hoisted(() => ({
  getMyInstances: vi.fn(),
  destroyInstance: vi.fn(),
  extendInstance: vi.fn(),
  requestInstanceAccess: vi.fn(),
}))

vi.mock('@/api/instance', () => instanceApiMocks)

describe('InstanceList', () => {
  beforeEach(() => {
    instanceApiMocks.getMyInstances.mockResolvedValue([
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
      {
        id: 'inst-2',
        challenge_id: 'chal-2',
        challenge_title: '反序列化迷宫',
        category: 'web',
        difficulty: 'medium',
        status: 'pending',
        access_url: '',
        flag_type: 'dynamic',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 1,
        created_at: '2026-03-05T00:00:00Z',
        queue_position: 2,
        eta_seconds: 90,
        progress: 35,
      },
    ])
  })

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
    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.text()).toContain('我的实例')
    expect(wrapper.text()).toContain('SQL 注入基础')
    expect(wrapper.text()).toContain('反序列化迷宫')
    expect(wrapper.text()).toContain('等待创建')
    expect(wrapper.text()).toContain('实例正在排队创建')
  })
})
