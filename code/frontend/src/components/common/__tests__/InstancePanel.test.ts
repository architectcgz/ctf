import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import InstancePanel from '../InstancePanel.vue'

describe('InstancePanel', () => {
  it('应该渲染实例列表', async () => {
    const wrapper = mount(InstancePanel, {
      props: {
        instances: [
          {
            id: '1',
            challenge_id: '1',
            challenge_title: 'Test Challenge',
            category: 'web',
            difficulty: 'easy',
            status: 'running',
            access_url: 'http://test.com',
            flag_type: 'static',
            share_scope: 'per_user',
            expires_at: new Date(Date.now() + 3600000).toISOString(),
            remaining_extends: 2,
            created_at: '2024-01-01T00:00:00Z',
          },
        ],
      },
    })

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('Test Challenge')
  })
})
