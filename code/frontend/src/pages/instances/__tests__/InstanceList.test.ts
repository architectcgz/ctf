import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import InstanceList from '@/pages/instances/InstanceListRoutePage.vue'

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
      {
        id: 'inst-3',
        challenge_id: 'chal-3',
        challenge_title: '文件下载审计',
        category: 'web',
        difficulty: 'hard',
        status: 'failed',
        access_url: 'http://127.0.0.1:39999',
        flag_type: 'dynamic',
        share_scope: 'per_user',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 0,
        created_at: '2026-03-05T00:00:00Z',
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
    expect(wrapper.text()).toContain('Instances')
    expect(wrapper.text()).toContain('我的实例')
    expect(wrapper.text()).toContain('SQL 注入基础')
    expect(wrapper.text()).toContain('反序列化迷宫')
    expect(wrapper.text()).toContain('文件下载审计')
    expect(wrapper.text()).toContain('等待创建')
    expect(wrapper.text()).toContain('实例正在排队创建')
    expect(wrapper.text()).toContain('启动失败，当前目标不可访问')
    expect(wrapper.text()).toContain('系统托管')
    expect(wrapper.find('.instance-row-title').attributes('title')).toBe('SQL 注入基础')
    expect(wrapper.find('.instance-row-access-value').attributes('title')).toBe(
      'http://example.test'
    )
    expect(wrapper.text()).not.toContain('http://127.0.0.1:39999')
  })

  it('AWD 队伍实例不应显示延时或销毁操作', async () => {
    instanceApiMocks.getMyInstances.mockResolvedValueOnce([
      {
        id: 'awd-inst-1',
        challenge_id: 'awd-service-1',
        challenge_title: 'Bank Portal',
        category: 'web',
        difficulty: 'medium',
        status: 'running',
        access_url: '',
        flag_type: 'dynamic',
        share_scope: 'per_team',
        contest_mode: 'awd',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 1,
        created_at: '2026-03-05T00:00:00Z',
      },
    ])

    const wrapper = mount(InstanceList, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await flushPromises()

    const row = wrapper.get('.instance-row')
    expect(row.text()).toContain('Bank Portal')
    expect(row.text()).toContain('系统托管')
    expect(row.text()).not.toContain('延时')
    expect(row.text()).not.toContain('销毁')
  })

  it('即将过期提醒应具备对话框语义、关闭按钮和 ESC 关闭能力', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-03-05T00:00:00Z'))
    instanceApiMocks.getMyInstances.mockResolvedValueOnce([
      {
        id: 'inst-expiring',
        challenge_id: 'chal-expiring',
        challenge_title: '即将过期靶机',
        category: 'web',
        difficulty: 'easy',
        status: 'running',
        access_url: 'http://example.test',
        flag_type: 'dynamic',
        share_scope: 'per_user',
        expires_at: '2026-03-05T00:04:00Z',
        remaining_extends: 1,
        created_at: '2026-03-05T00:00:00Z',
      },
    ])

    const wrapper = mount(InstanceList, {
      attachTo: document.body,
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await flushPromises()
    vi.advanceTimersByTime(1000)
    await flushPromises()

    const dialog = wrapper.get('[role="dialog"]')
    expect(dialog.attributes('aria-modal')).toBe('true')
    expect(dialog.attributes('aria-labelledby')).toBe('instance-warning-title')
    expect(dialog.attributes('aria-describedby')).toBe('instance-warning-description')
    expect(wrapper.text()).toContain('实例即将过期')
    expect(wrapper.get('button[aria-label="关闭实例过期提醒"]').element).toBe(
      document.activeElement
    )

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    wrapper.unmount()
    vi.useRealTimers()
  })

  it('等待创建中的实例应定时重新同步服务端状态', async () => {
    vi.useFakeTimers()
    instanceApiMocks.getMyInstances
      .mockResolvedValueOnce([
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
      .mockResolvedValueOnce([
        {
          id: 'inst-2',
          challenge_id: 'chal-2',
          challenge_title: '反序列化迷宫',
          category: 'web',
          difficulty: 'medium',
          status: 'running',
          access_url: 'http://instance.ready.test',
          flag_type: 'dynamic',
          share_scope: 'per_user',
          expires_at: '2099-01-01T00:00:00Z',
          remaining_extends: 1,
          created_at: '2026-03-05T00:00:00Z',
        },
      ])

    const wrapper = mount(InstanceList, {
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>',
          },
        },
      },
    })

    await flushPromises()
    expect(wrapper.text()).toContain('等待创建')

    vi.advanceTimersByTime(5000)
    await flushPromises()

    expect(instanceApiMocks.getMyInstances.mock.calls.length).toBeGreaterThanOrEqual(2)
    expect(wrapper.text()).toContain('运行中')
    expect(wrapper.text()).toContain('http://instance.ready.test')

    vi.useRealTimers()
  })
})
