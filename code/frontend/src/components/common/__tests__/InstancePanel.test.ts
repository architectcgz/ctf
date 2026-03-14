import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import InstancePanel from '../InstancePanel.vue'

vi.mock('@/api/instance', () => ({
  getMyInstances: vi.fn().mockResolvedValue([
    {
      id: '1',
      challenge_id: '1',
      challenge_title: 'Test Challenge',
      status: 'running',
      access_url: 'http://test.com',
      expires_at: new Date(Date.now() + 3600000).toISOString(),
      remaining_extends: 2,
      created_at: '2024-01-01T00:00:00Z',
    },
  ]),
  destroyInstance: vi.fn(),
  extendInstance: vi.fn(),
  requestInstanceAccess: vi.fn(),
}))

describe('InstancePanel', () => {
  it('应该渲染实例列表', async () => {
    const wrapper = mount(InstancePanel)

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.exists()).toBe(true)
  })
})
