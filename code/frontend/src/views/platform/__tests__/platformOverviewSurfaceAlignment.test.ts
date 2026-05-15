import { describe, expect, it } from 'vitest'

import adminDashboardSource from '@/components/platform/dashboard/PlatformOverviewPage.vue?raw'

describe('admin dashboard surface alignment', () => {
  it('softens the hero primary action border and focus ring to match the dark surface system', () => {
    expect(adminDashboardSource).toContain('class="header-actions overview-action-grid"')
    expect(adminDashboardSource).toContain('class="header-btn header-btn--primary"')
    expect(adminDashboardSource).toContain('class="header-btn header-btn--ghost"')
    expect(adminDashboardSource).not.toContain('overview-action-main')
    expect(adminDashboardSource).toMatch(
      /\.workspace-alert-actions\s*>\s*\.ui-btn\.ui-btn--ghost\s*\{[\s\S]*--ui-btn-border:\s*var\(--journal-border\);[\s\S]*--ui-btn-background:\s*color-mix\(in srgb,\s*var\(--journal-surface\) 94%, transparent\);/s
    )
  })

  it('frames the overview hero action rail as a compact operational panel', () => {
    expect(adminDashboardSource).toContain('class="overview-hero-actions"')
    expect(adminDashboardSource).toContain('class="header-actions overview-action-grid"')
    expect(adminDashboardSource).toContain('class="workspace-page-header overview-page-header"')
    expect(adminDashboardSource).toMatch(
      /\.overview-hero-actions\s*\{[\s\S]*border:\s*1px solid var\(--workspace-line-soft\);[\s\S]*border-radius:\s*var\(--workspace-radius-lg\);[\s\S]*background:/s
    )
    expect(adminDashboardSource).toMatch(
      /\.overview-hero-actions\s*\{[\s\S]*align-self:\s*start;[\s\S]*width:\s*min\(19rem,\s*100%\);[\s\S]*padding:\s*var\(--space-3\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.overview-page-header\s*\{[\s\S]*align-items:\s*start;/s
    )
    expect(adminDashboardSource).toMatch(
      /\.hero-meta-badge\s*\{[\s\S]*padding-bottom:\s*var\(--space-2\);[\s\S]*border-bottom:\s*1px solid var\(--workspace-line-soft\);/s
    )
    expect(adminDashboardSource).toMatch(
      /\.overview-action-grid\s*\{[\s\S]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\);/s
    )
  })

  it('uses shared admin summary cards and directory shells instead of page-private premium panels', () => {
    expect(adminDashboardSource).toContain(
      'class="admin-summary-grid overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(adminDashboardSource).toContain('class="journal-note progress-card metric-panel-card"')
    expect(adminDashboardSource).toContain('class="workspace-directory-section overview-directory-section"')
    expect(adminDashboardSource).toContain('class="workspace-directory-list overview-list-shell"')
    expect(adminDashboardSource).not.toContain('metric-panel-card--premium')
    expect(adminDashboardSource).not.toContain('metric-panel-grid--premium')
  })

  it('uses list-heading for alerts and hotspot section headers instead of workspace-tab-heading', () => {
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
