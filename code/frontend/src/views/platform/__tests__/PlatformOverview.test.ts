import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'

import adminDashboardPageSourceBase from '@/features/platform/overview/ui/PlatformOverviewPage.vue?raw'
import platformOverviewPageModelSource from '@/features/platform/overview/model/usePlatformOverviewPage.ts?raw'
import platformOverviewRoutesSource from '@/features/platform/overview/model/platformOverviewRoutes.ts?raw'
import platformOverviewAlertsSectionSource from '@/features/platform/overview/ui/PlatformOverviewAlertsSection.vue?raw'
import platformOverviewHeroPanelSource from '@/features/platform/overview/ui/PlatformOverviewHeroPanel.vue?raw'
import platformOverviewHotspotsSectionSource from '@/features/platform/overview/ui/PlatformOverviewHotspotsSection.vue?raw'
import PlatformOverview from '@/pages/platform/PlatformOverviewRoutePage.vue'
import platformOverviewViewSource from '@/pages/platform/PlatformOverviewRoutePage.vue?raw'

const adminDashboardPageSource = [
  adminDashboardPageSourceBase,
  platformOverviewHeroPanelSource,
  platformOverviewAlertsSectionSource,
  platformOverviewHotspotsSectionSource,
].join('\n')

const adminApiMocks = vi.hoisted(() => ({
  getDashboard: vi.fn(),
}))

vi.mock('@/api/admin/platform', () => adminApiMocks)

describe('PlatformOverview', () => {
  beforeEach(() => {
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

  function mountPage() {
    return mount(PlatformOverview, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })
  }

  it('应该展示系统概览指标与容器告警', async () => {
    const wrapper = mountPage()

    await flushPromises()

    expect(adminApiMocks.getDashboard).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('系统值守台')
    expect(wrapper.text()).toContain('18')
    expect(wrapper.text()).toContain('6')
    expect(wrapper.text()).toContain('CPU 持续高于阈值')
    expect(wrapper.text()).toContain('web-01')
  })

  it('CPU 占用低于 1% 时应保留小数显示，并给热点条保留最小宽度', async () => {
    adminApiMocks.getDashboard.mockResolvedValueOnce({
      online_users: 18,
      active_containers: 6,
      cpu_usage: 0.01,
      memory_usage: 47,
      container_stats: [
        {
          container_id: 'ctf-web-1',
          container_name: 'web-01',
          cpu_percent: 0.32,
          memory_percent: 54,
          memory_usage: 1073741824,
          memory_limit: 2147483648,
        },
      ],
      alerts: [],
    })

    const wrapper = mountPage()

    await flushPromises()

    expect(wrapper.text()).toContain('0.01%')
    expect(wrapper.text()).toContain('0.32%')

    const usageBars = wrapper.findAll('.usage-bar')
    expect(usageBars[0]?.attributes('style')).toContain('width: 1%;')
  })

  it('路由页应只做组合，不直接处理平台概览请求与导航', () => {
    expect(platformOverviewViewSource).toContain('usePlatformOverviewPage')
    expect(platformOverviewViewSource).not.toContain("from '@/api/admin/platform'")
    expect(platformOverviewViewSource).not.toContain("router.push({ name: 'AuditLog' })")
    expect(platformOverviewPageModelSource).not.toContain("from 'vue-router'")
    expect(platformOverviewPageModelSource).toContain('auditLogRoute: buildPlatformAuditLogRoute()')
    expect(platformOverviewRoutesSource).toContain('platformCheatDetectionRoute')
  })

  it('应该移除页面内 tab，并直接展示总览、当前告警与资源热点三个区块', async () => {
    const wrapper = mountPage()

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

  it('头部操作应改用共享 header-btn 原语而不是页面私有 admin-btn 按钮族', () => {
    expect(adminDashboardPageSource).toContain('class="header-btn header-btn--primary"')
    expect(adminDashboardPageSource).toContain('class="header-btn header-btn--ghost"')
    expect(adminDashboardPageSource).toContain(
      'class="header-btn header-btn--ghost overview-anchor-btn"'
    )
    expect(adminDashboardPageSource).not.toContain('overview-action-main')
    expect(adminDashboardPageSource).not.toContain('admin-btn admin-btn-primary')
    expect(adminDashboardPageSource).not.toContain('admin-btn admin-btn-ghost')
  })

  it('总览页应通过 route target contract 渲染审计与风险研判入口', async () => {
    const wrapper = mountPage()

    await flushPromises()

    const auditLink = wrapper
      .findAllComponents(RouterLinkStub)
      .find((link) => link.text().includes('审计日志'))
    const cheatLink = wrapper
      .findAllComponents(RouterLinkStub)
      .find((link) => link.text().includes('风险研判'))

    expect(auditLink?.props('to')).toEqual({ name: 'AuditLog' })
    expect(cheatLink?.props('to')).toEqual({ name: 'CheatDetection' })
  })

  it('应该采用与 teacher dashboard 一致的 workspace 骨架，并去掉页面内重复顶栏', () => {
    expect(adminDashboardPageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero overview-shell"'
    )
    expect(adminDashboardPageSource).not.toContain('class="workspace-topbar"')
    expect(adminDashboardPageSource).toContain('class="content-pane overview-content"')
    expect(adminDashboardPageSource).toContain('class="workspace-page-header overview-page-header"')
    expect(adminDashboardPageSource).not.toContain('class="workspace-hero"')
    expect(adminDashboardPageSource).not.toContain('tab-panel')
    expect(adminDashboardPageSource).toContain('class="hero-title workspace-page-title"')
    expect(adminDashboardPageSource).toContain('系统值守台')
    expect(adminDashboardPageSource).toContain('class="hero-summary workspace-page-copy"')
    expect(adminDashboardPageSource).toContain('class="meta-strip"')
    expect(adminDashboardPageSource).toContain(
      'class="admin-summary-grid overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(adminDashboardPageSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(adminDashboardPageSource).toContain('class="overview-hero-actions"')
    expect(adminDashboardPageSource).toContain('class="header-actions overview-action-grid"')
  })

  it('总览面板应将系统脉搏收进 hero 右侧操作轨道，而不是单独的 rail 区块', () => {
    expect(adminDashboardPageSource).not.toContain('overview-pulse-panel')
    expect(adminDashboardPageSource).not.toContain('class="hero-rail"')
    expect(adminDashboardPageSource).toContain('class="hero-meta-badge"')
    expect(adminDashboardPageSource).toContain('System Pulse')
  })

  it('总览 premium 指标条应使用数字列数变量，避免四个指标退化成单列', () => {
    expect(adminDashboardPageSource).toContain('--metric-panel-columns: 4;')
    expect(adminDashboardPageSource).not.toContain(
      '--metric-panel-columns: repeat(4, minmax(0, 1fr));'
    )
  })
})
