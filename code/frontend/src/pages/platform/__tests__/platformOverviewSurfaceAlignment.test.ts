import { describe, expect, it } from 'vitest'

import adminDashboardSourceBase from '@/features/platform/overview/ui/PlatformOverviewPage.vue?raw'
import platformOverviewAlertsSectionSource from '@/features/platform/overview/ui/PlatformOverviewAlertsSection.vue?raw'
import platformOverviewHeroPanelSource from '@/features/platform/overview/ui/PlatformOverviewHeroPanel.vue?raw'
import platformOverviewHotspotsSectionSource from '@/features/platform/overview/ui/PlatformOverviewHotspotsSection.vue?raw'

const adminDashboardSource = [
  adminDashboardSourceBase,
  platformOverviewHeroPanelSource,
  platformOverviewAlertsSectionSource,
  platformOverviewHotspotsSectionSource,
].join('\n')

describe('admin dashboard surface alignment', () => {
  it('platform overview should keep hero and sections on shared workspace owners', () => {
    expect(adminDashboardSource).toContain('workspace-page-header')
    expect(adminDashboardSource).toContain('workspace-directory-section')
    expect(adminDashboardSource).toContain('workspace-directory-list')
    expect(adminDashboardSource).not.toContain('metric-panel-card--premium')
    expect(adminDashboardSource).not.toContain('metric-panel-grid--premium')
  })

  it('platform overview should keep section heading owner on list-heading instead of workspace-tab-heading', () => {
    expect(adminDashboardSource).toContain('当前告警')
    expect(adminDashboardSource).toContain('资源热点')
    expect(adminDashboardSource).toContain('list-heading__title')
    expect(adminDashboardSource).not.toContain('workspace-tab-heading__title">当前告警</h2>')
    expect(adminDashboardSource).not.toContain('workspace-tab-heading__title">资源热点</h2>')
  })
})
