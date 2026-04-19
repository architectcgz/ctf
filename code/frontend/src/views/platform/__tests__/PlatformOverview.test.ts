import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import adminDashboardPageSource from '@/components/platform/dashboard/PlatformOverviewPage.vue?raw'
import PlatformOverview from '../PlatformOverview.vue'

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

describe('PlatformOverview', () => {
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
    const wrapper = mount(PlatformOverview)

    await flushPromises()

    expect(adminApiMocks.getDashboard).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('系统值守台')
    expect(wrapper.text()).toContain('18')
    expect(wrapper.text()).toContain('6')
    expect(wrapper.text()).toContain('CPU 持续高于阈值')
    expect(wrapper.text()).toContain('web-01')
  })

  it('应该移除页面内 tab，并直接展示总览、当前告警与资源热点三个区块', async () => {
    const wrapper = mount(PlatformOverview)

    await flushPromises()

    expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
    expect(wrapper.find('#admin-dashboard-tab-overview').exists()).toBe(false)
    expect(wrapper.find('#admin-dashboard-tab-alerts').exists()).toBe(false)
    expect(wrapper.find('#admin-dashboard-tab-hotspots').exists()).toBe(false)

    expect(wrapper.find('#admin-dashboard-overview').exists()).toBe(true)
    expect(wrapper.find('#admin-dashboard-alerts').exists()).toBe(true)
    expect(wrapper.find('#admin-dashboard-hotspots').exists()).toBe(true)
    expect(wrapper.find('#admin-dashboard-overview').text()).toContain('审计日志')
    expect(wrapper.find('#admin-dashboard-alerts').text()).toContain('CPU 持续高于阈值')
    expect(wrapper.find('#admin-dashboard-hotspots').text()).toContain('web-01')
  })

  it('应该去掉页面内顶部标签栏', () => {
    expect(adminDashboardPageSource).not.toContain('role="tablist"')
    expect(adminDashboardPageSource).not.toContain('class="top-tabs"')
    expect(adminDashboardPageSource).not.toContain('admin-dashboard-tab-overview')
    expect(adminDashboardPageSource).not.toContain('admin-dashboard-tab-alerts')
    expect(adminDashboardPageSource).not.toContain('admin-dashboard-tab-hotspots')
  })

  it('应改用共享 ui-btn 原语而不是页面私有 admin-btn 按钮族', () => {
    expect(adminDashboardPageSource).toContain('class="quick-action ui-btn ui-btn--primary"')
    expect(adminDashboardPageSource).toContain('class="quick-action ui-btn ui-btn--ghost"')
    expect(adminDashboardPageSource).not.toContain('admin-btn admin-btn-primary')
    expect(adminDashboardPageSource).not.toContain('admin-btn admin-btn-ghost')
  })

  it('应该采用与 teacher dashboard 一致的 workspace 骨架，并去掉页面内重复顶栏', () => {
    expect(adminDashboardPageSource).toContain('class="workspace-shell"')
    expect(adminDashboardPageSource).not.toContain('class="workspace-topbar"')
    expect(adminDashboardPageSource).toContain('class="content-pane"')
    expect(adminDashboardPageSource).toContain('class="workspace-hero"')
    expect(adminDashboardPageSource).not.toContain('tab-panel')
    expect(adminDashboardPageSource).toMatch(/class="[^"]*\bhero-title\b[^"]*"/)
    expect(adminDashboardPageSource).toContain('系统值守台')
    expect(adminDashboardPageSource).toContain('class="hero-summary"')
    expect(adminDashboardPageSource).toContain('class="meta-strip"')
    expect(adminDashboardPageSource).toMatch(/class="[^"]*\bprogress-strip\b[^"]*"/)
    expect(adminDashboardPageSource).toContain('class="hero-rail"')
  })

  it('总览面板不应再保留额外的 pulse article 区块', () => {
    expect(adminDashboardPageSource).not.toContain('overview-pulse-panel')
    expect(adminDashboardPageSource).not.toContain('运行脉搏')
  })
})
