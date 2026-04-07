import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import adminDashboardPageSource from '@/components/admin/dashboard/AdminDashboardPage.vue?raw'
import AdminDashboard from '../AdminDashboard.vue'

const pushMock = vi.fn()

const adminApiMocks = vi.hoisted(() => ({
  getDashboard: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)

describe('AdminDashboard', () => {
  beforeEach(() => {
    pushMock.mockReset()
    adminApiMocks.getDashboard.mockReset()
    adminApiMocks.getDashboard.mockResolvedValue({
      online_users: 18,
      active_containers: 6,
      cpu_usage: 62,
      memory_usage: 47,
      container_stats: [
        {
          container_id: 'ctf-web-1',
          container_name: 'web-01',
          cpu_percent: 71,
          memory_percent: 54,
          memory_usage: 1073741824,
          memory_limit: 2147483648,
        },
      ],
      alerts: [
        {
          container_id: 'ctf-web-1',
          type: 'cpu',
          value: 91,
          threshold: 80,
          message: 'CPU 持续高于阈值',
        },
      ],
    })
  })

  it('应该展示系统概览指标与容器告警', async () => {
    const wrapper = mount(AdminDashboard)

    await flushPromises()

    expect(adminApiMocks.getDashboard).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('系统值守台')
    expect(wrapper.text()).toContain('18')
    expect(wrapper.text()).toContain('6')
    expect(wrapper.text()).toContain('CPU 持续高于阈值')
    expect(wrapper.text()).toContain('web-01')
  })

  it('应该将总览、当前告警与资源热点拆分为独立 tab', async () => {
    const wrapper = mount(AdminDashboard)

    await flushPromises()

    expect(wrapper.find('#admin-dashboard-tab-overview').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#admin-dashboard-tab-alerts').attributes('aria-selected')).toBe('false')
    expect(wrapper.find('#admin-dashboard-tab-hotspots').attributes('aria-selected')).toBe('false')

    expect(wrapper.find('#admin-dashboard-panel-overview').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#admin-dashboard-panel-alerts').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#admin-dashboard-panel-hotspots').attributes('aria-hidden')).toBe('true')
    expect(wrapper.find('#admin-dashboard-panel-overview').text()).toContain('审计日志')

    await wrapper.get('#admin-dashboard-tab-alerts').trigger('click')

    expect(wrapper.find('#admin-dashboard-tab-alerts').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#admin-dashboard-panel-alerts').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#admin-dashboard-panel-alerts').text()).toContain('CPU 持续高于阈值')

    await wrapper.get('#admin-dashboard-tab-hotspots').trigger('click')

    expect(wrapper.find('#admin-dashboard-tab-hotspots').attributes('aria-selected')).toBe('true')
    expect(wrapper.find('#admin-dashboard-panel-hotspots').attributes('aria-hidden')).toBe('false')
    expect(wrapper.find('#admin-dashboard-panel-hotspots').text()).toContain('web-01')
  })

  it('应该将标签栏放在页面顶部标题之前，与其他工作台页面保持一致', () => {
    expect(adminDashboardPageSource).toContain('role="tablist"')
    expect(adminDashboardPageSource.indexOf('role="tablist"')).toBeLessThan(
      adminDashboardPageSource.indexOf('系统值守台'),
    )
  })

  it('应该采用与 teacher dashboard 一致的 workspace 骨架，而不是独立的中后台容器风格', () => {
    expect(adminDashboardPageSource).toContain('class="workspace-shell"')
    expect(adminDashboardPageSource).toContain('class="workspace-topbar"')
    expect(adminDashboardPageSource).toContain('class="content-pane"')
    expect(adminDashboardPageSource).toContain('class="workspace-hero tab-panel"')
    expect(adminDashboardPageSource).toContain('class="hero-title"')
    expect(adminDashboardPageSource).toContain('系统值守台')
    expect(adminDashboardPageSource).toContain('class="hero-summary"')
    expect(adminDashboardPageSource).toContain('class="meta-strip"')
    expect(adminDashboardPageSource).toContain('class="progress-strip"')
    expect(adminDashboardPageSource).toContain('class="hero-rail"')
  })

  it('总览面板不应再保留额外的 pulse article 区块', () => {
    expect(adminDashboardPageSource).not.toContain('overview-pulse-panel')
    expect(adminDashboardPageSource).not.toContain('运行脉搏')
  })
})
