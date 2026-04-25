import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import InstancePanel from '../InstancePanel.vue'
import instancePanelSource from '../InstancePanel.vue?raw'

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

  it('实例面板应接入共享表面与按钮原语，而不是继续依赖 Element Plus 卡片和按钮', () => {
    expect(instancePanelSource).toContain('class="instance-panel"')
    expect(instancePanelSource).toContain('class="ui-btn ui-btn--secondary ui-btn--sm"')
    expect(instancePanelSource).toContain('class="ui-btn ui-btn--primary ui-btn--sm"')
    expect(instancePanelSource).not.toContain('<ElCard')
    expect(instancePanelSource).not.toContain('<ElButton')
    expect(instancePanelSource).not.toContain('<ElTag')
  })

  it('倒计时状态颜色应通过语义类承接，而不是从函数返回 Tailwind 任意主题类', () => {
    expect(instancePanelSource).toContain('instance-countdown--danger')
    expect(instancePanelSource).not.toContain('text-[var(--color-danger)]')
    expect(instancePanelSource).not.toContain('text-[var(--color-warning)]')
    expect(instancePanelSource).not.toContain('text-[var(--color-success)]')
    expect(instancePanelSource).not.toContain('text-[var(--color-text-muted)]')
  })
})
