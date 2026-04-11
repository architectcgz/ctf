import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import InstanceList from '../InstanceList.vue'
import instanceListSource from '../InstanceList.vue?raw'

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
        share_scope: 'shared',
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
        share_scope: 'per_user',
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
    expect(wrapper.text()).toContain('Instances')
    expect(wrapper.text()).toContain('我的实例')
    expect(wrapper.text()).toContain('SQL 注入基础')
    expect(wrapper.text()).toContain('反序列化迷宫')
    expect(wrapper.text()).toContain('等待创建')
    expect(wrapper.text()).toContain('实例正在排队创建')
    expect(wrapper.text()).toContain('系统托管')
    expect(wrapper.find('.instance-row-title').attributes('title')).toBe('SQL 注入基础')
    expect(wrapper.find('.instance-row-access-value').attributes('title')).toBe('http://example.test')
  })

  it('应该为实例列表长标题和访问地址保留省略样式与完整提示', () => {
    expect(instanceListSource).toMatch(/class="instance-row-title"[\s\S]*:title="instance\.challenge_title"/s)
    expect(instanceListSource).toMatch(/instance-row-access-value[\s\S]*:title="/s)
    expect(instanceListSource).toMatch(/\.instance-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s)
    expect(instanceListSource).toMatch(/\.instance-row-access-value\s*\{[^}]*display:\s*-webkit-box;[^}]*-webkit-line-clamp:\s*2;[^}]*overflow:\s*hidden;/s)
  })

  it('实例页概况卡片应使用统一 metric-panel 样式类', () => {
    expect(instanceListSource).toContain('class="instance-summary-grid metric-panel-grid"')
    expect(instanceListSource).toContain('class="instance-summary-item metric-panel-card"')
    expect(instanceListSource).toContain('class="instance-summary-label metric-panel-label"')
    expect(instanceListSource).toContain('class="instance-summary-value metric-panel-value"')
    expect(instanceListSource).toContain('class="instance-summary-helper metric-panel-helper"')
    expect(instanceListSource).toContain('当前仍在运行、可直接访问的实例数量')
    expect(instanceListSource).toContain('已经提交创建请求、正在排队或启动中的实例数量')
    expect(instanceListSource).toContain('当前账号最多可同时保留的实例数量')
  })
})
