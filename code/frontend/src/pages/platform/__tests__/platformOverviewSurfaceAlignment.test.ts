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
  it('platform overview should keep hero action owner on shared header actions instead of page-private actions', () => {
    expect(adminDashboardSource).toContain('class="overview-hero-actions"')
    expect(adminDashboardSource).toContain('class="header-actions overview-action-grid"')
    expect(adminDashboardSource).toContain('class="workspace-page-header overview-page-header"')
    expect(adminDashboardSource).toContain('class="header-btn header-btn--primary"')
    expect(adminDashboardSource).toContain('class="header-btn header-btn--ghost"')
    expect(adminDashboardSource).not.toContain('overview-action-main')
  })

  it('platform overview should keep summary and directory owner on shared workspace primitives', () => {
    expect(adminDashboardSource).toContain('class="workspace-directory-section overview-directory-section"')
    expect(adminDashboardSource).toContain('class="workspace-directory-list overview-list-shell"')
    expect(adminDashboardSource).not.toContain('metric-panel-card--premium')
    expect(adminDashboardSource).not.toContain('metric-panel-grid--premium')
  })

  it('platform overview should keep section heading owner on list-heading instead of workspace-tab-heading', () => {
    expect(adminDashboardSource).toContain(
      '<h2 class="section-title list-heading__title">当前告警</h2>'
    )
    expect(adminDashboardSource).toContain(
      '<h2 class="section-title list-heading__title">资源热点</h2>'
    )
    expect(adminDashboardSource).not.toContain('workspace-tab-heading__title">当前告警</h2>')
    expect(adminDashboardSource).not.toContain('workspace-tab-heading__title">资源热点</h2>')
  })
})
